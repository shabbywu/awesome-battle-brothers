package matches

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/heroiclabs/nakama-common/runtime"
	"online-championships/pkg/matches/strategicProperties"
	"online-championships/pkg/protocol"
	"strconv"
	"time"
)

type MatchState struct {
	factions                map[int]*Presence
	presences               map[string]*Presence
	totalFaction            int
	battleResult            map[int]bool
	orderedPresences        []string
	battleStarted           bool
	battleEnd               bool
	battleEndEventBroadcast bool
	strategicProperties     strategicProperties.StrategicProperties
	features                map[int]bool
	roomRuntime             *RoomRuntime
	joinSeats               map[string]protocol.Seat
	joinReservations        map[string]JoinReservation
	roomID                  string
}

type Match struct{}

const DefaultJoinReservationTimeoutTicks int64 = 30

type JoinReservation struct {
	Seat         protocol.Seat
	Presence     SeatPresence
	AcceptedTick int64
	DeadlineTick int64
}

func NewMatch(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule) (m runtime.Match, err error) {
	return &Match{}, nil
}

func (m *Match) MatchInit(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, params map[string]interface{}) (interface{}, int, string) {
	state := &MatchState{
		factions:         map[int]*Presence{},
		presences:        map[string]*Presence{},
		battleResult:     map[int]bool{},
		features:         map[int]bool{},
		roomRuntime:      NewRoomRuntime(""),
		joinSeats:        map[string]protocol.Seat{},
		joinReservations: map[string]JoinReservation{},
	}

	createdParams := MatchCreateParams{}
	_ = json.Unmarshal([]byte(params["createdParams"].(string)), &createdParams)
	state.strategicProperties = strategicProperties.BuildStrategicProperties(createdParams.Scenario, createdParams.MapSeed, createdParams.CombatSeed)

	tickRate := MatchTickRate
	name := "EmptyRoom"
	if createdParams.Name != "" {
		name = createdParams.Name
		state.features[FeatureCustomMatchName] = true
	}
	description := createdParams.Description
	metadata := buildRoomMetadata(
		createdParams,
		name,
		description,
		state.strategicProperties.MapSeed,
		state.strategicProperties.CombatSeed,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	state.roomRuntime.Metadata = metadata
	label := string(DumpEvent(MatchLabel{Name: name, Description: description, Metadata: metadata}))
	return state, tickRate, label
}

func (m *Match) MatchJoinAttempt(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presence runtime.Presence, metadata map[string]string) (interface{}, bool, string) {
	mState, _ := state.(*MatchState)
	requestedSeat := protocol.SeatAuto
	if seat := metadata["seat"]; seat != "" {
		requestedSeat = protocol.Seat(seat)
	}
	if matchID := metadata["matchId"]; matchID != "" && mState.roomID == "" {
		mState.roomID = matchID
	}
	seat, err := mState.reserveJoinSeat(toSeatPresence(presence), requestedSeat, tick)
	if err != nil {
		logger.Info("join rejected for %s: %v", presence.GetSessionId(), err)
		return mState, false, err.Error()
	}
	mState.joinSeats[presence.GetSessionId()] = seat
	return mState, true, ""
}

func (m *Match) MatchJoin(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)
	mState.expireJoinReservations(tick)

	for _, p := range presences {
		seat, err := mState.commitJoinSeat(toSeatPresence(p))
		if err != nil {
			logger.Error("unable to assign seat for %s: %v", p.GetSessionId(), err)
			if kickErr := dispatcher.MatchKick([]runtime.Presence{p}); kickErr != nil {
				logger.Error("kick unseated presence error: %v", kickErr)
			}
			continue
		}
		mState.presences[p.GetSessionId()] = spawnPlayer(p)
		mState.orderedPresences = append(mState.orderedPresences, p.GetSessionId())
		if err := dispatcher.BroadcastMessageDeferred(protocol.OpCodeSeatAssigned, DumpEvent(protocol.SeatAssignedEvent{
			SchemaVersion: protocol.SchemaVersion,
			Seat:          seat,
			UserID:        p.GetUserId(),
			SessionID:     p.GetSessionId(),
		}), []runtime.Presence{p}, nil, true); err != nil {
			logger.Error("broadcast seat assignment error: %v", err)
		}
		m.replayRoomToPresence(logger, dispatcher, mState, p)
	}
	m.broadcastRoomState(logger, dispatcher, mState)
	return mState
}

func (m *Match) MatchLeave(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, presences []runtime.Presence) interface{} {
	mState, _ := state.(*MatchState)

	combatResult := CombatResultNone
	for _, p := range presences {
		o := mState.presences[p.GetSessionId()]
		if mState.roomRuntime != nil {
			mState.roomRuntime.RemovePresenceAt(p.GetSessionId(), tick)
			delete(mState.joinSeats, p.GetSessionId())
			delete(mState.joinReservations, p.GetSessionId())
		}
		if o != nil && o.Faction != FactionNone {
			if mState.battleStarted {
				combatResult = convertMatchResult(combatResult, o.Faction)
			}
		}
		delete(mState.presences, p.GetSessionId())
	}

	// rebuild orderedPresences
	var orderedPresences []string
	for _, o := range mState.orderedPresences {
		if _, ok := mState.presences[o]; ok {
			orderedPresences = append(orderedPresences, o)
		}
	}
	mState.orderedPresences = orderedPresences

	if mState.battleStarted && !mState.battleEnd {
		switch combatResult {
		case CombatResultBlueSideWin:
			mState.battleResult[FactionBlueSide] = true
			mState.battleResult[FactionRedSide] = false
			mState.battleEnd = true
		case CombatResultRedSideWin:
			mState.battleResult[FactionBlueSide] = false
			mState.battleResult[FactionRedSide] = true
			mState.battleEnd = true
		case CombatResultDraw:
			mState.battleResult[FactionBlueSide] = false
			mState.battleResult[FactionRedSide] = false
			mState.battleEnd = true
		default:
			break
		}
	}
	if mState.roomRuntime != nil {
		if len(mState.presences) > 0 {
			m.broadcastRoomState(logger, dispatcher, mState)
		}
		if len(mState.presences) == 0 && shouldKeepRuntimeWithoutPresence(mState.roomRuntime.Status) {
			return mState
		}
	}
	if len(mState.presences) == 0 {
		return nil
	}
	return mState
}

func (m *Match) MatchLoop(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, messages []runtime.MatchData) interface{} {
	mState, _ := state.(*MatchState)
	mState.expireJoinReservations(tick)

	for _, message := range messages {
		if handle, ok := protocolHandlers[message.GetOpCode()]; ok {
			handle(ctx, logger, db, nk, dispatcher, state, message, tick)
		} else if handle, ok := handlers[message.GetOpCode()]; ok {
			if mState.roomRuntime != nil {
				logger.Info("rejecting legacy opcode in v1 room: %d", message.GetOpCode())
				handleProtocolError(logger, dispatcher, protocol.ProtocolError{
					Code:    protocol.ErrorInvalidCommand,
					Message: fmt.Sprintf("legacy opcode not allowed in v1 room: %d", message.GetOpCode()),
				}, 0, nil, message, message.GetOpCode(), "")
			} else {
				handle(ctx, logger, db, nk, dispatcher, state, message)
			}
		} else {
			logger.Info("dropping unknown match opcode: %d", message.GetOpCode())
			handleProtocolError(logger, dispatcher, protocol.ProtocolError{
				Code:    protocol.ErrorUnknownOpcode,
				Message: fmt.Sprintf("unknown match opcode: %d", message.GetOpCode()),
			}, 0, nil, message, message.GetOpCode(), "")
		}
	}

	checkRoomRuntimeTimeout(logger, dispatcher, mState, tick)
	checkRoomRuntimeBattleEndReportTimeout(logger, dispatcher, mState, tick)
	checkRoomRuntimeAbandon(logger, dispatcher, mState, tick)

	if mState.roomRuntime == nil {
		if !mState.battleStarted {
			m.balanceFaction(logger, state, dispatcher)
		} else if mState.battleEnd && mState.totalFaction > 0 && len(mState.battleResult) == mState.totalFaction && !mState.battleEndEventBroadcast {
			m.broadcastBattleResult(logger, state, dispatcher)
		}
	}

	return mState
}

func checkRoomRuntimeTimeout(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState, tick int64) {
	if state.roomRuntime == nil {
		return
	}
	serverSeq, err := state.roomRuntime.ExpirePendingCheckpoints(tick)
	if err == nil {
		return
	}
	handleProtocolError(logger, dispatcher, err, serverSeq, nil, nil, protocol.OpCodeCmdResult, "")
	broadcastRoomState(logger, dispatcher, state)
}

func checkRoomRuntimeBattleEndReportTimeout(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState, tick int64) {
	if state.roomRuntime == nil {
		return
	}
	if err := state.roomRuntime.ExpireBattleEndReports(tick); err == nil {
		return
	} else {
		handleProtocolError(logger, dispatcher, err, 0, nil, nil, protocol.OpCodeMatchBattleEndReport, "")
		broadcastRoomState(logger, dispatcher, state)
	}
}

func checkRoomRuntimeAbandon(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState, tick int64) {
	if state.roomRuntime == nil {
		return
	}
	result, expired, err := state.roomRuntime.ExpireAbandonedSeats(state.roomID, tick)
	if err != nil {
		handleProtocolError(logger, dispatcher, err, 0, nil, nil, protocol.OpCodeMatchForfeitRequest, "")
		return
	}
	if !expired {
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeMatchResult, DumpEvent(result), nil, nil, true); err != nil {
		logger.Error("broadcast abandon match result error: %v", err)
	}
	broadcastRoomState(logger, dispatcher, state)
}

func (m *Match) MatchTerminate(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, graceSeconds int) interface{} {
	message := "Server shutting down in " + strconv.Itoa(graceSeconds) + " seconds."
	// _ = dispatcher.BroadcastMessage(OpCodeMatchTerminate, []byte(message), nil, nil, true)
	logger.Info(message)
	return state
}

func (m *Match) MatchSignal(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, dispatcher runtime.MatchDispatcher, tick int64, state interface{}, data string) (interface{}, string) {
	var request MatchSignalRequest
	if err := json.Unmarshal([]byte(data), &request); err != nil {
		return state, string(DumpEvent(matchSignalError(MatchSignalKindCommandLogQuery, protocol.ErrorInvalidCommand, "unable to unmarshal signal payload")))
	}
	switch request.Kind {
	case MatchSignalKindCommandLogQuery:
		if request.CommandLog == nil {
			return state, string(DumpEvent(matchSignalError(request.Kind, protocol.ErrorInvalidCommand, "missing commandLog query")))
		}
		response, err := commandLogQueryResponse(state, *request.CommandLog)
		if err != nil {
			return state, string(DumpEvent(matchSignalError(request.Kind, protocol.ErrorCodeOf(err), err.Error())))
		}
		return state, string(DumpEvent(response))
	default:
		return state, string(DumpEvent(matchSignalError(request.Kind, protocol.ErrorInvalidCommand, "unsupported match signal kind")))
	}
}

func matchSignalError(kind string, code protocol.ErrorCode, message string) MatchSignalErrorResponse {
	if code == "" {
		code = protocol.ErrorInvalidCommand
	}
	return MatchSignalErrorResponse{
		SchemaVersion: protocol.SchemaVersion,
		Kind:          kind,
		Status:        "error",
		Code:          code,
		Message:       message,
	}
}

func commandLogQueryResponse(state interface{}, query CommandLogQueryRequest) (CommandLogQueryResponse, error) {
	mState, _ := state.(*MatchState)
	if mState == nil || mState.roomRuntime == nil {
		return CommandLogQueryResponse{}, protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "match runtime is not available"}
	}
	if query.MatchID == "" {
		return CommandLogQueryResponse{}, protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "matchId is required"}
	}
	if mState.roomID != "" && query.MatchID != mState.roomID {
		return CommandLogQueryResponse{}, protocol.ProtocolError{Code: protocol.ErrorInvalidCommand, Message: "matchId does not match running room"}
	}

	room := mState.roomRuntime
	fromSeq := query.FromServerSeq
	if fromSeq <= 0 {
		fromSeq = 1
	}
	limit := query.Limit
	if limit <= 0 {
		limit = DefaultCommandLogQueryLimit
	}
	if limit > MaxCommandLogQueryLimit {
		limit = MaxCommandLogQueryLimit
	}

	response := CommandLogQueryResponse{
		SchemaVersion:         protocol.SchemaVersion,
		Kind:                  MatchSignalKindCommandLogQuery,
		MatchID:               query.MatchID,
		Status:                "ok",
		Commands:              []protocol.CommandEnvelope{},
		FromServerSeq:         fromSeq,
		NextServerSeq:         room.NextServerSeq,
		LastCompletedSeq:      room.LastCompletedSeq(),
		PendingServerSeq:      room.PendingServerSeq(),
		CommandLogLength:      len(room.CommandLog),
		CompletedResultCount:  len(room.CompletedResults),
		MaxLimit:              MaxCommandLogQueryLimit,
		Persistence:           "in-memory",
		CanPromoteMVP:         false,
		CanPromotePersistence: false,
	}
	if query.IncludeRoomState {
		roomState := room.RoomState(query.MatchID)
		response.RoomState = &roomState
	}
	if query.IncludeStartSnapshot && room.StartSnapshot != nil {
		startSnapshot := *room.StartSnapshot
		response.StartSnapshot = &startSnapshot
	}
	if query.IncludeCompletedResults {
		response.CompletedResults = []protocol.CommandResult{}
	}

	for _, command := range room.CommandLog {
		if command.ServerSeq < fromSeq {
			continue
		}
		if len(response.Commands) >= limit {
			response.HasMore = true
			response.NextFromServerSeq = command.ServerSeq
			break
		}
		response.Commands = append(response.Commands, command)
		if query.IncludeCompletedResults {
			if result, ok := room.CompletedResult(command.ServerSeq); ok {
				response.CompletedResults = append(response.CompletedResults, result)
			}
		}
	}
	if !response.HasMore && len(response.Commands) > 0 {
		response.NextFromServerSeq = response.Commands[len(response.Commands)-1].ServerSeq + 1
	} else if !response.HasMore {
		response.NextFromServerSeq = fromSeq
	}

	if query.IncludePending {
		pendingSeq := room.PendingServerSeq()
		if pendingSeq > 0 {
			if pending, ok := room.PendingCheckpoints[pendingSeq]; ok {
				command := pending.Command
				response.PendingCommand = &command
			}
		}
	}
	if query.IncludeTerminalResult && room.TerminalResult != nil {
		terminalResult := *room.TerminalResult
		response.TerminalResult = &terminalResult
	}
	return response, nil
}

func (m *Match) balanceFaction(logger runtime.Logger, state interface{}, dispatcher runtime.MatchDispatcher) {
	mState, _ := state.(*MatchState)
	if len(mState.factions) == 2 {
		return
	}

	var players []string
	for _, faction := range []int{FactionRedSide, FactionBlueSide} {
		if p, used := mState.factions[faction]; used {
			players = append(players, p.GetUsername())
			continue
		}
		for _, sessionId := range mState.orderedPresences {
			p := mState.presences[sessionId]
			if p.Faction == FactionNone {
				p.Faction = faction
				mState.factions[faction] = p
				players = append(players, p.GetUsername())

				// broadcast balance event
				err := dispatcher.BroadcastMessageDeferred(OpCodeMatchFactionDispatch, DumpEvent(FactionDispatchEvent{
					Presence: sessionId,
					Faction:  faction,
				}), nil, nil, true)
				if err != nil {
					logger.Error("broadcast error: %v", err)
				}
			}
		}
	}
	if !mState.features[FeatureCustomMatchName] {
		var name string
		switch len(players) {
		case 1:
			name = fmt.Sprintf("vs %s", players[0])
			break
		case 2:
			name = fmt.Sprintf("%s vs %s", players[0], players[1])
			break
		default:
			name = "EmptyRoom"
		}
		metadata := defaultRoomMetadata()
		if mState.roomRuntime != nil {
			metadata = mState.roomRuntime.Metadata
		}
		metadata.Name = name
		metadata.Description = ""
		if mState.roomRuntime != nil {
			mState.roomRuntime.Metadata = metadata
		}
		label := string(DumpEvent(MatchLabel{Name: name, Description: "", Metadata: metadata}))
		if err := dispatcher.MatchLabelUpdate(label); err != nil {
			logger.Error("MatchLabelUpdate error: %v", err)
		}
	}
}

func toSeatPresence(presence runtime.Presence) SeatPresence {
	return SeatPresence{
		UserID:    presence.GetUserId(),
		SessionID: presence.GetSessionId(),
		Username:  presence.GetUsername(),
	}
}

func (s *MatchState) reserveJoinSeat(presence SeatPresence, requested protocol.Seat, tick int64) (protocol.Seat, error) {
	if s.joinSeats == nil {
		s.joinSeats = map[string]protocol.Seat{}
	}
	if s.joinReservations == nil {
		s.joinReservations = map[string]JoinReservation{}
	}
	s.expireJoinReservations(tick)

	if !protocol.IsJoinSeat(requested) {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(requested)}
	}
	if s.roomRuntime.Status == protocol.RoomStatusClosed {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(s.roomRuntime.Status)}
	}

	seat, err := s.previewJoinSeat(presence, requested)
	if err != nil {
		return "", err
	}
	s.joinReservations[presence.SessionID] = JoinReservation{
		Seat:         seat,
		Presence:     presence,
		AcceptedTick: tick,
		DeadlineTick: joinReservationDeadline(tick),
	}
	s.joinSeats[presence.SessionID] = seat
	return seat, nil
}

func (s *MatchState) previewJoinSeat(presence SeatPresence, requested protocol.Seat) (protocol.Seat, error) {
	if existing, ok := s.roomRuntime.findSeatByUserID(presence.UserID); ok {
		return existing, nil
	}
	if requested == protocol.SeatSpectator {
		return s.previewSpectatorJoinSeat(presence)
	}
	if s.roomRuntime.Status == protocol.RoomStatusEnded {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(s.roomRuntime.Status)}
	}
	if s.roomRuntime.Status != protocol.RoomStatusOpen && s.roomRuntime.Status != protocol.RoomStatusReadyCheck {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: string(s.roomRuntime.Status)}
	}
	if requested == protocol.SeatAuto {
		if !s.playerSeatReserved(protocol.SeatRed, presence.SessionID) {
			requested = protocol.SeatRed
		} else {
			requested = protocol.SeatBlue
		}
	}
	if !protocol.IsPlayerSeat(requested) {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: string(requested)}
	}
	if s.playerSeatReserved(requested, presence.SessionID) {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidSeat, Message: fmt.Sprintf("%s is occupied", requested)}
	}
	return requested, nil
}

func (s *MatchState) previewSpectatorJoinSeat(presence SeatPresence) (protocol.Seat, error) {
	if _, ok := s.roomRuntime.findSpectatorSessionByUserID(presence.UserID); ok {
		return protocol.SeatSpectator, nil
	}
	if !s.roomRuntime.Metadata.Spectators.Enabled {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "spectators disabled"}
	}
	if s.roomRuntime.Metadata.Spectators.Limit > 0 &&
		s.reservedSpectatorCount(presence.SessionID) >= s.roomRuntime.Metadata.Spectators.Limit {
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "spectator limit reached"}
	}
	return protocol.SeatSpectator, nil
}

func (s *MatchState) playerSeatReserved(seat protocol.Seat, sessionID string) bool {
	if _, occupied := s.roomRuntime.Seats[seat]; occupied {
		return true
	}
	for reservationSessionID, reservation := range s.joinReservations {
		if reservationSessionID == sessionID {
			continue
		}
		if reservation.Seat == seat {
			return true
		}
	}
	return false
}

func (s *MatchState) reservedSpectatorCount(sessionID string) int {
	userIDs := map[string]bool{}
	for _, presence := range s.roomRuntime.Spectators {
		userIDs[presence.UserID] = true
	}
	for reservationSessionID, reservation := range s.joinReservations {
		if reservationSessionID == sessionID || reservation.Seat != protocol.SeatSpectator {
			continue
		}
		userIDs[reservation.Presence.UserID] = true
	}
	return len(userIDs)
}

func (s *MatchState) commitJoinSeat(presence SeatPresence) (protocol.Seat, error) {
	if s.joinReservations == nil {
		s.joinReservations = map[string]JoinReservation{}
	}
	reservation, ok := s.joinReservations[presence.SessionID]
	if !ok {
		delete(s.joinSeats, presence.SessionID)
		return "", protocol.ProtocolError{Code: protocol.ErrorInvalidState, Message: "join reservation expired or missing"}
	}

	seat, err := s.roomRuntime.AssignSeat(presence, reservation.Seat)
	delete(s.joinReservations, presence.SessionID)
	delete(s.joinSeats, presence.SessionID)
	return seat, err
}

func (s *MatchState) expireJoinReservations(tick int64) {
	if len(s.joinReservations) == 0 {
		return
	}
	for sessionID, reservation := range s.joinReservations {
		if reservation.DeadlineTick == 0 || tick < reservation.DeadlineTick {
			continue
		}
		delete(s.joinReservations, sessionID)
		delete(s.joinSeats, sessionID)
	}
}

func joinReservationDeadline(tick int64) int64 {
	return tick + DefaultJoinReservationTimeoutTicks
}

func shouldKeepRuntimeWithoutPresence(status protocol.RoomStatus) bool {
	return status == protocol.RoomStatusInBattle ||
		status == protocol.RoomStatusPaused ||
		status == protocol.RoomStatusDesynced
}

func (m *Match) replayRoomToPresence(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState, presence runtime.Presence) {
	if state.roomRuntime == nil {
		return
	}
	target := []runtime.Presence{presence}
	if err := dispatcher.BroadcastMessageDeferred(protocol.OpCodeRoomState, DumpEvent(state.roomRuntime.RoomState(state.roomID)), target, nil, true); err != nil {
		logger.Error("broadcast room state replay error: %v", err)
	}
	if state.roomRuntime.StartSnapshot == nil {
		return
	}
	if err := dispatcher.BroadcastMessageDeferred(protocol.OpCodeMatchStart, DumpEvent(*state.roomRuntime.StartSnapshot), target, nil, true); err != nil {
		logger.Error("broadcast match start replay error: %v", err)
	}
	for _, cmd := range state.roomRuntime.CommandLog {
		result, completed := state.roomRuntime.CompletedResult(cmd.ServerSeq)
		if !completed {
			if _, pending := state.roomRuntime.PendingCheckpoints[cmd.ServerSeq]; pending {
				replayCmd := cmd
				replayCmd.Replay = true
				replayCmd.ReplayRequiresResult = replayRequiresResultFromPresence(state.roomRuntime, presence, cmd.ServerSeq)
				opCode, ok := protocol.OpcodeForCommand(replayCmd.Op)
				if !ok {
					logger.Error("unable to replay unsupported pending command op: %s", replayCmd.Op)
					break
				}
				if err := dispatcher.BroadcastMessageDeferred(opCode, DumpEvent(replayCmd), target, nil, true); err != nil {
					logger.Error("broadcast pending command replay error: %v", err)
				}
			}
			break
		}
		replayCmd := cmd
		replayCmd.Replay = true
		opCode, ok := protocol.OpcodeForCommand(replayCmd.Op)
		if !ok {
			logger.Error("unable to replay unsupported command op: %s", replayCmd.Op)
			continue
		}
		if err := dispatcher.BroadcastMessageDeferred(opCode, DumpEvent(replayCmd), target, nil, true); err != nil {
			logger.Error("broadcast command replay error: %v", err)
		}
		replayResult := result
		replayResult.Replay = true
		if err := dispatcher.BroadcastMessageDeferred(protocol.OpCodeCmdResult, DumpEvent(replayResult), target, nil, true); err != nil {
			logger.Error("broadcast command result replay error: %v", err)
		}
	}
	if state.roomRuntime.TerminalResult != nil {
		replayMatchResult := *state.roomRuntime.TerminalResult
		replayMatchResult.Replay = true
		if err := dispatcher.BroadcastMessageDeferred(protocol.OpCodeMatchResult, DumpEvent(replayMatchResult), target, nil, true); err != nil {
			logger.Error("broadcast terminal match result replay error: %v", err)
		}
	}
}

func replayRequiresResultFromPresence(room *RoomRuntime, presence runtime.Presence, serverSeq int64) bool {
	seat, ok := room.SeatForSession(presence.GetSessionId())
	if !ok || !protocol.IsPlayerSeat(seat) {
		return false
	}
	pending, ok := room.PendingCheckpoints[serverSeq]
	if !ok {
		return false
	}
	_, hasResult := pending.Results[seat]
	return !hasResult
}

func (m *Match) broadcastRoomState(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState) {
	broadcastRoomState(logger, dispatcher, state)
}

func broadcastRoomState(logger runtime.Logger, dispatcher runtime.MatchDispatcher, state *MatchState) {
	if state.roomRuntime == nil {
		return
	}
	if len(state.presences) == 0 {
		return
	}
	if err := dispatcher.BroadcastMessage(protocol.OpCodeRoomState, DumpEvent(state.roomRuntime.RoomState(state.roomID)), nil, nil, true); err != nil {
		logger.Error("broadcast room state error: %v", err)
	}
}

func (m *Match) broadcastBattleResult(logger runtime.Logger, state interface{}, dispatcher runtime.MatchDispatcher) {
	mState, _ := state.(*MatchState)
	mState.battleEndEventBroadcast = true
	if mState.roomRuntime != nil {
		mState.roomRuntime.Status = protocol.RoomStatusEnded
	}
	var event BattleEndEvent
	for f, IsVictory := range mState.battleResult {
		if IsVictory {
			event.IsVictory = IsVictory
			event.Faction = f
		}
	}

	if err := dispatcher.BroadcastMessage(OpCodeBattleEnd, DumpEvent(event), nil, nil, true); err != nil {
		logger.Error("broadcast error: %v", err)
	}
}
