Awesome Battle Brothers(震惊！战场兄弟)项目

Awesome Battle Brothers 项目旨在增强游戏对 Mod 的支持和体验。

[中文说明](./README.md) | [EN README](./README_EN.md)

---
⭐️ 启动器 - bb-launcher
==========================================
bb-launcher 是 Battle Brothers(战场兄弟) 游戏的增强启动器, 除了能正常游玩游戏以外，增加了以下新功能:
- 内置 4GB 补丁, 减缓游戏内存不足的问题 (包括 Steam！)
- 支持自定义大地图渲染和 UI 的字体 (汉化项目需要依赖该功能)
- 支持使用 [webkit 浏览器](https://www.nexusmods.com/battlebrothers/mods/744) 进行远程调试 JS/UI

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
 
⭐️ 联机Mod服务端 - online-championships-server
==========================================
online-championships-server 是 Battle Brothers(战场兄弟) 游戏联机对战 Mod 的服务端程序, 该模块基于 Nakama Server 开发, 提供了以下功能:
1. 联机对战
2. ...
