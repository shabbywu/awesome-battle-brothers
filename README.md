Awesome Battle Brothers(震惊！战场兄弟)项目

Awesome Battle Brothers 项目旨在增强游戏对 Mod 的支持和体验。

[中文说明](./README.md) | [EN README](./README_EN.md)

---
⭐️ 启动器 - bb-launcher
==========================================
bb-launcher 是 Battle Brothers(战场兄弟) 游戏的增强启动器, 除了能正常游玩游戏以外，增加了以下新功能:
- 内置 4GB 补丁, 减缓游戏内存不足的问题 (包括 Steam！)
- 支持自定义大地图渲染和 UI 的字体 (汉化项目需要依赖该功能)
- 支持自定义 Mod 加载逻辑(顺序和路径)！

使用说明
==========================================
1. 从 Release 中下载附件 **bb-launcher-steam.exe.zip** 或 **bb-launcher-gog.exe.zip**
2. 解压缩后将 **.exe** 文件放在游戏根目录或 win32 目录, (游戏根目录通常是 **X:\\Program Files (x86)\\Steam\\steamapps\\common\\Battle Brothers\\**)
3. Steam 版本的启动器需要确保**Steam 已在后台运行**且使用 Steam 启动 1 次游戏已触发 Steam 的云联网更新机制
4. 运行 .exe 启动游戏

配置文件说明
==========================================
初次启动游戏后, 会根据当前的 Mod 列表在游戏根目录下生成配置文件 `awesome-battle-brothers\settings.toml`, 通过调整配置文件可实现 **自定义字体** 和 **自定义 Mod 加载逻辑**。

### 关于自定义字体(Font)
1. 选择合适的字体, 将字体文件的绝对路径配置到 Font, 即可使用自定义字体
2. Windows 用户可在 **C:\Windows\Fonts** 目录下找到系统内置的字体
 
### 关于自定义 Mod 加载顺序(Mods)
1. 当新增 Mod 时, 在使用启动器时会自动更新 Mod 列表
2. 当 Mod 不存在或加载失败时, 启动器会跳过异常的 Mod
3. 设置 enabled = false 可以跳过加载该 Mod
4. 调整 order 值可以控制 Mod 的加载顺序
5. 除非你正在游戏目录外开发 Mod, 否则请勿随意调整 location 值


配置文件内容(示例)如下:
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
