# 贪吃蛇项目架构文档

嗨，欢迎接手这个贪吃蛇项目！这篇文档会带你快速了解整个项目是怎么搭起来的，看完应该就能上手改代码了，不用一行行猜意图。

---

## 整体架构一句话

这是一个纯 Go 写的 Windows 命令行贪吃蛇，采用经典的**单线程轮询 + Ticker 定时驱动**模型：`main()` 跑一个死循环，每次先轮询键盘输入，再检查定时触发的游戏步进，最后根据脏标记决定要不要重绘画面。各个模块（输入、游戏逻辑、渲染、存档）完全解耦，通过 `Game` 结构体作为中枢串起来。

---

## 各文件职责速览

| 文件 | 一句话概括 |
|---|---|
| [main.go](../main.go) | 程序入口，初始化资源、跑主循环、清理资源 |
| [game.go](../game.go) | 游戏大脑，管状态、管碰撞、管协调各个模块 |
| [snake.go](../snake.go) | 蛇的领域对象：身体、方向、移动、碰撞检测 |
| [food.go](../food.go) | 食物的领域对象：随机生成、被吃判定 |
| [obstacle.go](../obstacle.go) | 障碍物的领域对象：生成规则（避开蛇初始位置）、碰撞检测 |
| [render.go](../render.go) | 画面渲染：控制台清屏、画地图、画侧边栏、PAUSED 覆盖层 |
| [keyboard.go](../keyboard.go) | 键盘输入：Windows 控制台原始模式，方向键/R/Q/空格 |
| [score.go](../score.go) | 最高分存档：读写 `highscore.txt` |
| [go.mod](../go.mod) | Go 模块声明，零第三方依赖 |

---

## 核心数据结构关系

### Game 是总指挥官

`Game` 结构体是整个游戏的中枢，它**组合**了所有领域对象，不继承、不派生，纯组合：

```
Game
├─ Snake      // 蛇（身体+方向）
├─ Food       // 食物（位置）
├─ Obstacles  // 障碍物列表
├─ Score / HighScore / SpeedLevel  // 数值状态
├─ GameOver / Paused / Running     // 状态标记
├─ Ticker     // 时间驱动
└─ NeedRender // 脏标记 → 要不要重绘
```

**它们之间怎么通信？**
- `Game` **不直接操作** Snake/Food/Obstacles 的内部字段，都是调用它们的方法。比如 `game.Step()` 里调用 `snake.Move()` 拿新头位置，再调用 `snake.CollidesWall()`、`obstacles.Has()` 检查碰撞。
- Snake/Food/Obstacles **互相不知道对方的存在**，所有跨对象的逻辑（比如"生成食物要避开蛇和障碍物"）都放在 `Game` 或 `Food` 的方法参数里传进去。
- 状态变更统一走 `Game` 的方法：`TogglePause()`、`Restart()`、`endGame()`、`HandleKey()`，这些方法内部会调用 `MarkDirty()` 告诉渲染层"我变了，下次该重绘了"。

### 渲染层怎么解耦？

`render.go` 定义了一个 `CellSource` 接口：
```go
type CellSource interface {
    GetCellContent(x, y int) string
}
```
对外的 `Render()` 函数签名保持不变（给 `main.go` 调用），内部会把参数包成 `mapCellSource` 适配器，核心渲染逻辑只通过接口拿数据，不直接依赖 Snake/Food/Obstacles 的具体结构。以后想改数据结构，只要改适配器就行。

---

## 主循环执行流程

### 伪代码长这样

```go
for game.Running {
    game.PollInput()         // 1. 轮询键盘，更新状态+设脏标

    select {
    case <-game.Ticker.C:    // 2. 时间到了？
        if game.IsActive() {
            game.Step()      // 3. 蛇走一步、检测碰撞、吃食物
        } else {
            game.MarkDirty() // 4. 哪怕不走，状态变了也要重绘
        }
    default:
    }

    if game.NeedRender {     // 5. 脏了就重绘
        Render(...)
        game.NeedRender = false
    }
}
```

### 时序图（输入 → 状态 → 渲染 完整链路）

```
 用户按键
    │
    ▼
ReadKeyEvent()  [keyboard.go]
    │  解析出 KeyEvent{Action, Direction}
    ▼
Game.HandleKey()  [game.go]
    ├─ ActionPause → TogglePause() → 停Ticker/重启Ticker → MarkDirty()
    ├─ ActionRestart → Restart() → 重置所有状态 → MarkDirty()
    ├─ ActionQuit → Running=false → MarkDirty()
    └─ ActionDir → 活动状态才改方向 → MarkDirty()
    │
    ▼
Ticker.C 触发（按速度档）
    │
    ▼
Game.Step()  [game.go]
    ├─ 非活动状态直接返回
    ├─ snake.Move() → 算新头位置
    ├─ checkCollision() → 撞墙/撞自己/撞障碍 → endGame()
    ├─ 吃食物 → snake.Advance(, grow=true) + 加分 + 刷食物
    └─ 没吃 → snake.Advance(, grow=false)
    │
    ▼
NeedRender = true
    │
    ▼
Render()  [render.go]
    └─ renderWithSource(CellSource) → 逐格查询 → 画地图 → 画侧边栏 → 暂停就盖个PAUSED框
```

**关键点**：输入和 Ticker 是两个独立的事件源，**都通过 MarkDirty() 触发重绘**，没有谁等谁的问题。

---

## 关键函数速查

### main.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `main()` | 入口：初始化键盘/渲染 → 跑主循环 → 清理 | 操作系统 |

### game.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `NewGame()` | 创建新游戏，初始化蛇/障碍物/食物/分数 | `main()`, `Restart()` |
| `IsActive()` | `!GameOver && !Paused`，能不能走 | `Step()`, `main()` |
| `TogglePause()` | 切暂停，会真的停/启 Ticker | `HandleKey()` |
| `Restart()` | 重开，重置所有状态+刷障碍物 | `HandleKey()` |
| `HandleKey()` | 分发键盘事件 | `PollInput()` |
| `Step()` | 游戏走一帧 | `main()`（Ticker 触发） |
| `endGame()` | 结束游戏，存高分，停 Ticker | `Step()` |
| `PollInput()` | 轮询键盘直到没输入 | `main()` |
| `GetCellContent()` | 给渲染层用，返回某格的字符 | `mapCellSource` 适配器 |
| `MarkDirty()` | 设 `NeedRender=true` | 所有改状态的地方 |

### snake.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `NewSnake()` | 创建初始蛇（3节，朝右） | `NewGame()`, `Restart()` |
| `Move()` | 算下一个头位置，**不真的动** | `Step()` |
| `Advance()` | 真的移动，grow 控制是否变长 | `Step()` |
| `CollidesWall()` | 撞墙检测 | `checkCollision()` |
| `CollidesSelf()` | 撞自己检测 | `checkCollision()` |
| `Occupies()` | 某位置有没有蛇身 | 渲染、食物生成 |
| `IsHead()` | 某位置是不是蛇头 | 渲染 |

### food.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `Spawn()` | 在空位随机刷食物（避开蛇+障碍物） | `NewGame()`, `Restart()`, `Step()` |
| `EatenBy()` | 蛇头是不是吃到食物了 | `Step()` |

### obstacle.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `NewObstacles()` | 生成 3~5 个障碍物，避开蛇初始位置 2 格 | `NewGame()`, `Restart()` |
| `Has()` | 某位置是不是障碍物 | `checkCollision()`、渲染、食物生成 |

### render.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `Render()` | 对外入口，参数包成 CellSource | `main()` |
| `renderWithSource()` | 核心渲染逻辑，只认 CellSource 接口 | `Render()` |
| `ClearScreen()` | 清屏 | `main()` |

### keyboard.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `ReadKeyEvent()` | 读一次键盘，不阻塞 | `PollInput()` |

### score.go

| 函数 | 作用 | 被谁调用 |
|---|---|---|
| `LoadHighScore()` | 从 highscore.txt 读最高分 | `NewGame()` |
| `UpdateHighScoreIfNeeded()` | 破纪录就写文件 | `endGame()` |

---

## 新手扩展指南

想加新功能？看这里找对地方改：

### 我想加个新道具（比如无敌星、减速药水）
- **数据结构**：仿照 `Food` 新建一个 PowerUp 类型，加在 `Game` 结构体里
- **生成逻辑**：在 `Game.Step()` 里加触发条件（比如每吃3个食物随机刷一个），调用 `PowerUp.Spawn()`，记得避开蛇和障碍物
- **效果逻辑**：吃到后改 `Game` 里的状态字段（比如 `Invincible bool`），在 `checkCollision()` 里判断无敌就跳过碰撞检测
- **渲染**：`Game.GetCellContent()` 里加个分支返回新字符（比如 `★`）
- **记得**：`Restart()` 里也要重置这个新状态

### 我想改游戏规则（比如时间限制、反向控制）
- 规则逻辑都在 `game.go` 里改：加 `TimeLeft` 字段、在 `Step()` 里倒计时、到 0 就 `endGame()`
- 反向控制的话，在 `HandleKey()` 里收到 `ActionDir` 后把方向取反再传给 `snake.SetDirection()`

### 我想换渲染方式（比如换成 Ebiten 图形界面）
- `render.go` 全换掉就行，对外只要提供同样签名的 `Render()` 函数
- 或者更干净：新写一个 `render_ebiten.go`，用 build tag 切换，`CellSource` 接口可以直接复用

### 我想加个音效
- 新建 `sound.go`，在 `endGame()`、吃食物、切暂停这些地方加调用
- 别忘 `Restart()` 里也要停掉正在播的音效

### 注意事项
1. **改状态一定要调用 `MarkDirty()`**，不然画面不会更新。
2. **`TogglePause()` 里真的会停 Ticker**，暂停期间不会有 tick 事件，不要在暂停时假设 Ticker 还在跑。
3. **对外函数签名别乱动**：`main.go` 调的 `Render()`、`keyboard.go` 的 `ReadKeyEvent()` 这些签名保持不变，不然调用方要跟着改。
4. **Windows 专用**：键盘和渲染都是用 Windows Console API 写的，跨平台的话要换实现。
5. **零第三方依赖**：加新库之前想想有没有必要，保持 `go.mod` 干净。

---

## 跑起来

```powershell
cd project98
go run .
```

操作：方向键移动，空格暂停/继续，R 重开，Q 退出。

祝你玩得开心，改得顺利！有问题翻源码，代码里注释不多但命名都尽量说人话了 😄
