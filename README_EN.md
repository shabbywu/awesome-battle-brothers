Awesome Battle Brothers(震惊！战场兄弟)项目

The Awesome Battle Brothers project aims to enhance the game's mod support and experience.

[中文说明](./README.md) | [EN README](./README_EN.md)

---
⭐️ Enhanced Launcher - bb-launcher
==========================================
The bb-launcher is an enhanced launcher for the game Battle Brothers. In addition to allowing you to play the game normally, it adds the following new features:

- Built-in 4GB patch to alleviate memory issues within the game (including Steam!)
- Supports custom rendering of the world map and UI fonts (necessary for localization projects)
- Supports Remote [UI Inspector](https://www.nexusmods.com/battlebrothers/mods/744)

Usage
==========================================
1. Download the attachment **bb-launcher-steam.exe.zip** or **bb-launcher-gog.exe.zip** from the Release.
2. After extracting, place the **.exe** file in the game’s root directory or win32 directory (the game’s root directory is usually **X:\\Program Files (x86)\\Steam\\steamapps\\common\\Battle Brothers\\**).
3. For the Steam version of the launcher, ensure that **Steam is running in the background** and that you have launched the game once through Steam to trigger Steam’s cloud update mechanism.
4. Run the .exe to launch the game.


Configuration File Instructions
==========================================
After launching the game for the first time, a configuration file `awesome-battle-brothers\settings.toml` will be generated in the game’s root directory based on the current list of mods. By adjusting this configuration file, you can customize **fonts** and **mod loading logic**.

### Custom Fonts (Font)
1. Select an appropriate font and configure the absolute path of the font file to the Font setting to use a custom font.
2. Windows users can find built-in system fonts in the **C:\Windows\Fonts** directory.

⭐️ Online Battle Mod Serverside - online-championships-server
==========================================
online-championships-server is the server program of the online battle mod of Battle Brothers. This module is developed based on Nakama Server and provides the following functions:
1. Online battle
2. ...