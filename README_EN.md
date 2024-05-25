Awesome Battle Brothers(震惊！战场兄弟)项目

Awesome Battle Brothers 项目旨在增强游戏对 Mod 的支持和体验。

[中文说明](./README.md) | [EN README](./README_EN.md)

---
⭐️ Enhanced Launcher - bb-launcher
==========================================
The bb-launcher is an enhanced launcher for the game Battle Brothers. In addition to allowing you to play the game normally, it adds the following new features:

- Built-in 4GB patch to alleviate memory issues within the game (including Steam!)
- Supports custom rendering of the world map and UI fonts (necessary for localization projects)
- Supports custom Mod loading logic (order and location)!

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

### Custom Mod Loading Logic (Mods)
1. When a new mod is added, the mod list will be automatically updated after running the launcher.
2. If a mod does not exist or fails to load, the launcher will skip the problematic mod.
3. Setting `enabled = false` will skip loading that mod.
4. Adjusting the `order` value can control the mod loading order.
5. Unless you are developing mods outside the game directory, do not adjust the `location` value.


Configuration File Content (Example) is as follows:
```toml
# Font: Optional[str]
# - Description: Absolute path to font file
# - 描述: 字体文件的绝对路径
# - Example(示例):
#
# Font = "C:\\Game\\data\\gfx\\fonts\\cinzel\\Cinzel-Black.ttf"
Font = ""

# Mods: List[str]
# - Description: List of Mods, which will be loaded in the order declared. Items are absolute path to mod file(***.data or ***.zip) or mod folder(the "data" folder),
# - 描述: Mods 列表, 将按声明的顺序加载. 每一项分别是 mod 文件（***.data 或 ***.zip）或 mod 文件夹（“data”文件夹）的绝对路径
# - Example(示例):
# [[Mods]]
# location = "C:\Game\data\data_001.dat"
# enabled = true
# order = 1
# [[Mods]]
# location = "C:\Game\data\"
# enabled = true
# order = 99999
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\Event Frequency 100-82-1-0-1605735549.zip"
order = 1
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\Lair Info Compilation-359-1-0-1613879637.zip"
order = 2
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\Named Item Rarity in Shops - 50 percent more chance-321-1-0-1604504297.zip"
order = 3
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\No Restriction Arena - V2-276-1-4-0-41-1601095752.zip"
order = 4
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\Sight_range_3x_hooked-78-1-4-0-40-1598538924.zip"
order = 5
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_001.dat"
order = 6
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_003.dat"
order = 7
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_004.dat"
order = 8
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_006.dat"
order = 9
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_008.dat"
order = 10
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_010.dat"
order = 11
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\data_160.dat"
order = 12
[[Mods]]
enabled = true
#📄 this mod is file
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data\\zdata_cn.zip"
order = 100000
[[Mods]]
enabled = true
#📂 this mod is directory
location = "D:\\SteamLibrary\\steamapps\\common\\Battle Brothers\\data"
order = 100001
```
