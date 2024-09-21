#pragma once
#include <bb-launcher-driver.h>
#include <squirrel.h>

void sqext_register_jsonlib(HSQUIRRELVM &v);
void sqext_register_msgpacklib(HSQUIRRELVM &v);
void sqext_register_bufferio(HSQUIRRELVM &v);
void sqext_override_isDevmode(HSQUIRRELVM &v);
void sqext_register_online_battle(HSQUIRRELVM &v);
void sqext_register_random(HSQUIRRELVM &v);
void sqext_register_print(HSQUIRRELVM &v);

namespace nakama {
void setup_client();
void fetch_tick();
} // namespace nakama

namespace online_battle {
bool mount_mod_to_physfs(launcher::LauncherMetadata *metadata);
}
