# goal-1 任务清单（B1：T-01 + T-02）

规则：一轮只做第一个未完成的 task。状态取值：`[ ]` 未开始 / `[x]` 完成 / `[!]` 阻塞（写明原因和需要的输入）。每完成 3 个普通 task 后安排一次 REVIEW。复查追加的修复 task（FIX-n）一经登记，就优先于排在它后面的普通 task 执行。
回写区每项都要填：**实际做了什么 / 验证证据（命令 + 结果 + commit）/ 剩余风险 / 下一步**。

---

## T-01 契约、工具链与构建基线

### [x] G1-01 工具链锁定与生成入口
- 依据：PLAN T-01 第 1 步；plan.md 假设 1、9。
- 内容：通过 `go get -tool google.golang.org/protobuf/cmd/protoc-gen-go` 固定生成器版本；新增 `contracts/generate.go`（`//go:generate`，核对 protoc 版本为 36.0）和一个占位 `.proto`（`contracts/proto/alonabot/v1/version.proto`，声明协议版本常量所在的包），能够生成并编译。把实际版本写入 docs/development.md 的环境表。
- 完成条件：`go generate ./...` 连续执行两次后 `git status` 不出现变化；`go build ./... && go vet ./...` 退出码为 0；protoc 版本不一致时生成明确失败（临时修改期望版本手工验证一次，再改回）。
- 回写：
  - 实际做了什么：`go get -tool` 固定 protoc-gen-go v1.36.12（go.mod 中 `tool` 指令 + `require google.golang.org/protobuf v1.36.12`）；新增 `contracts/doc.go`（`//go:generate go run ./internal/protocgen`）和 `contracts/internal/protocgen/main.go`（核对 `protoc --version` == `libprotoc 36.0` → 删除旧的 `pb/*.pb.go` → 用 `go tool -n protoc-gen-go` 得到的插件路径执行 protoc，参数为 `--go_opt=module=github.com/AlonaBot/alonabot/contracts`）。**与原计划不同：** 没有写占位的 version.proto，而是直接用 WIRE §2 已完整定义的 `errors.proto`（ErrorCode 0～17、ErrorInfo 1～3）验证整条流水线，避免留下无用文件；G1-02 不再需要写 errors.proto，只需补它的字段号测试。docs/development.md 的环境表已写入实际版本。
  - 验证证据（darwin/arm64，go1.27.1，libprotoc 36.0，本 task 的 commit）：`go generate ./...` 生成 `contracts/pb/alonabotv1/errors.pb.go`；`go mod tidy && go build ./... && go vet ./...` 退出码为 0；`git add -A` 后再连续执行两次 `go generate ./...`，`git diff --stat` 为空（结果稳定）；把期望版本临时改成 99.0 后，`go generate ./...` 以 exit 1 失败，报错为「protoc 版本为 "libprotoc 36.0"，需要 "libprotoc 99.0"」，且原有生成文件保留，随后已还原。
  - 剩余风险：版本核对用的是完整字符串比较，protoc 升级时必须同时修改 `wantProtoc` 并重新生成；生成器依赖 PATH 中的 protoc（homebrew），没有纳入 go.mod。目前还没有 go test（没有测试文件），自动化的生成稳定性检查在 REVIEW/CI 阶段处理。
  - 下一步：G1-02，写 envelope.proto 和 control.proto，并为 errors/envelope/control 编写字段号断言测试。

### [x] G1-02 Envelope、ErrorInfo 与控制帧 proto
- 依据：WIRE-FORMAT §1～§3、CONTRACTS §1.1。
- 内容：`envelope.proto`（Envelope 字段 1～16；InitiatorSide、FrameKind（含 SESSION_ERROR=13）、PayloadEncoding；字段 6/7 使用 optional）；`errors.proto` 已在 G1-01 完成，这里只补字段号测试；`control.proto`（Hello、HelloAck、Bootstrap（含 DatabaseDescriptor/ActivationReason、PhaseLimits、HttpProxyAuth）、Register、HttpListener、Prepared、Activate、Activated、Cancel、Stopping、Heartbeat、SessionError（failed_phase/failed_kind）、Capability）。payload 和 error 不放进 oneof。
- 完成条件：用 protoreflect 编写测试，逐项断言 Envelope 字段号、每个枚举的数值和名称前缀与 WIRE 表一致，optional 字段具备 presence；生成稳定。
- 回写：
  - 实际做了什么：新增 `envelope.proto`（Envelope 1～16，字段 6/7 为 optional，payload 和 error 不放进 oneof；InitiatorSide、FrameKind 0～13、PayloadEncoding）和 `control.proto`（ProtocolCapability、Hello、HelloAck、Bootstrap、DatabaseDescriptor、PostgresConnection、HttpProxyAuth、PhaseLimits、HttpListener、Prepared、Activate、Activated、Cancel、Stopping、Heartbeat、SessionError，以及枚举 DatabaseBackend（含显式 NONE）、MigrationMode、ActivationReason、CancelReason、StopReason、SessionPhase）。新增 `contracts/wirespec_test.go`：逐项断言 Envelope/ErrorInfo 的字段号与名称、presence（只有 6/7/16 有）、不放进 oneof；InitiatorSide/FrameKind/PayloadEncoding/ErrorCode 的数值和名称与 WIRE 表一致；三个文件中的所有枚举零值都是 `*_UNSPECIFIED` 且带类型前缀。
  - **规格之外由 T-01 决定的部分**（WIRE 只描述了语义，没有给出字段号或取值）：控制帧载荷的字段号按 WIRE §3.1/§3.2 表格中的字段顺序编号；CancelReason、StopReason、SessionPhase 的取值依据 CONTRACTS §4.2 中列出的阶段和停止原因；DatabaseBackend 增加了 NONE（对应无业务 DB 的插件，因为零值 UNSPECIFIED 必须被拒绝）。G1-03 会把这些写回 WIRE「T-01 已固定」小节。**REGISTER 载荷挪到 G1-03 的 route.proto**，因为它依赖 RouteDescriptor。
  - 验证证据（darwin/arm64，go1.27.1，libprotoc 36.0，本 task 的 commit）：`go generate ./... && go build ./... && go vet ./... && go test -race ./...` → `ok github.com/AlonaBot/alonabot/contracts`；stage 之后再执行 `go generate ./...`，`git diff --stat` 为空。变异检查：把 trace_id 的字段号临时改成 17 → `TestEnvelopeFieldNumbers` 失败（"Envelope 字段 11 应为 trace_id，实际 <nil>"），还原后通过。
  - 剩余风险：控制帧载荷由我自行编号，规格评审之前仍可能调整，但一旦冻结就只能 reserved，不能复用编号；`SessionError.failed_kind` 允许 UNSPECIFIED（表示与具体帧无关），G1-04 的校验需要对它单独处理。
  - 下一步：G1-03（route/Register、invocation、message、media、bus），并在 WIRE 中写「T-01 已固定」小节。

### [x] G1-03 业务载荷 proto：路由、调用、消息、回执、媒体、广播
- 依据：WIRE §3.3/§4/§5，MESSAGE-MODEL §1～§5，CONTRACTS §1.3/§1.4/§2。
- 内容：`route.proto`（REGISTER 载荷 Register：routes、config_schema_hash、candidate_config_json、candidate_config_version、可选 HttpListener、迁移完成信息；RouteDescriptor 公共字段，加 command/event/hook/service/schedule/http/bus_subscription 各自的描述，以及参数描述）、`invocation.proto`（可信 Invocation 上下文、HandlerOutcome、PreEventResult、BeforeSend/AfterSend 载荷）、`message.proto`（Event、Actor、Target、Capability、Segment 的各类型与 entities、MessageChain、SendResult、PartResult 及状态枚举）、`media.proto`（BeginUpload/WriteChunk/CommitUpload/Read/Grant 请求与响应、MediaRef、IncomingMedia）、`bus.proto`（BusEvent、Publish）。冻结字段号，把编号表写回 WIRE-FORMAT 新增的「T-01 已固定」小节，并在 adr/DECISIONS.md 的变更记录中加一行。
- 完成条件：全部 .proto 生成并编译通过；测试断言关键枚举的零值为 UNSPECIFIED，SendResult/PartResult 状态与 MESSAGE-MODEL 一致；规格文档同步更新。
- 回写：
  - 实际做了什么：新增 `identity.proto`（Target、Actor/TrustedRole、Capability；拆成单独文件是为了避免 message↔media 互相 import）、`message.proto`（Event + EventPayload oneof，含 8 类事件载荷；PlatformMessage；MessageChain/Segment 的 8 类片段，MediaSource 三选一，TextEntity 的 UTF-16 偏移；SendRequest、SendResult、PartResult、SendStatus、PartStatus）、`media.proto`（MediaRef/MediaScope、IncomingMedia/MediaReadability，以及 Begin/WriteChunk/Commit/Read/Grant 的请求与响应）、`route.proto`（Register + MigrationState；RouteDescriptor 公共字段加 7 类 spec oneof；Command/Event/Hook/Service/Schedule/Http/BusSubscription 描述与相关枚举）、`invocation.proto`（InvocationContext 含 send_forbidden 与 plugin_chain；HandlerRequest/CommandInvocation/ArgValue；HandlerOutcome；PreEvent/BeforeSend/AfterSend 的请求与结果）、`bus.proto`（PublishRequest/Response、BusEvent）。测试新增 SendStatus/PartStatus/Propagation/RouteKind/HookPhase 取值对照、HandlerOutcome 字段检查；枚举零值检查改为遍历 protoregistry 中 `alonabot.v1` 包的全部文件（至少 20 个枚举，否则直接判定失败）。spec 同步：WIRE-FORMAT 新增 §5A「T-01 已固定」，说明 proto 是编号的权威来源，并列出 T-01 做出的实现选择；adr/DECISIONS.md 变更记录加一行。
  - 验证证据（darwin/arm64，go1.27.1，libprotoc 36.0，本 task 的 commit）：`go generate ./...` 生成 9 个 `.pb.go`；`go build ./... && go vet ./...` 通过；`go test -race -v ./contracts/` 中 5 个测试全部 PASS；stage 之后再执行 `go generate ./...`，`git diff --stat` 为空。
  - 剩余风险：规格未定义而由我补定的结构（ForwardSource、RuleDescriptor 只有 has_custom_rule/require_chat_admin 两项、HttpAuthMode 的三种回调验证方式、Button.value 用 JSON 文本）在 T-03～T-07 实际使用时可能需要调整，所以 REVIEW-1 要逐字段对照规格。Core 固定服务名（例如 send、media.*、bus.publish 对应的 Envelope.target 与 schema_id）尚未定义，留给 G1-04（schema_id 映射）和 T-02/T-03。
  - 下一步：REVIEW-1。

### [x] REVIEW-1 集中检查（G1-01～G1-03）
- 检查清单：是否偏离 input.md；proto 与 WIRE/MESSAGE 逐字段对照（不能漏字段，也不能多出没有依据的字段）；生成是否稳定；`go vet`、`go build`；有无调试残留；文档是否同步。
- 发现问题就在本文件末尾追加修复 task。
- 回写：
  - 检查过程与结论（commit e9954e8 之上）：
    1. 是否偏离 input.md：没有偏离（只在 goal-1 分支内改动，依赖仍只有 protobuf v1.36.12，没有 push）。
    2. proto 与规格逐字段对照：WIRE §1/§2 的编号由测试锁定；§3.1 控制帧、§3.2 Bootstrap、§3.3 各 kind 的描述、MESSAGE-MODEL §1 Event/Actor/Target、§2 的 8 类 Segment、§3 HandlerOutcome、§4 SendResult/PartResult、§5 Hook 载荷、CONTRACTS §1.3/§1.4 的媒体请求逐项都能找到对应字段。规格表格之外新增的字段都有依据：EventSpec.action_id（MESSAGE §2：按钮 action_id 指向 button.clicked 路由）、ServiceSpec.encoding（CONTRACTS §3：描述包含编码）、IncomingMedia.expires_at（CONTRACTS §1.4：token 15 分钟有效）、BusEvent.encoding（服务与广播都支持 JSON/Protobuf 两种编码）。**发现缺口：** Service 调用（Core→目标插件）时，Envelope.payload 装的是服务的业务载荷，可信上下文（actor、scope、plugin_chain、send_forbidden）没有地方承载，违反 WIRE §4「可信调用上下文随 Invocation 载荷下发」→ 登记 FIX-1。
    3. 生成稳定：`go generate ./...` 没有 protoc 告警（无未使用的 import），之后 `git status` 为空。
    4. `go vet ./... && go build ./... && go test -race -count=1 ./...` 通过；`gofmt -l` 只列出 experiments/ 下的文件（独立 module，不属于本 goal）。
    5. 调试残留：没有 TODO/FIXME/println；protocgen 中的 `fmt.Fprintln(os.Stderr, ...)` 是有意的错误输出。
    6. 文档同步：ARCHITECTURE.md 的「当前状态」仍写着「只有 go.mod 和规格」，contracts 的状态还是「拟建」；docs/development.md 规则表中「Envelope 字段号/枚举值与 WIRE-FORMAT 一致」仍为「已定义，未实现」→ 一并放进 FIX-1。
    7. 安全与数据一致性：本阶段只有类型定义，不涉及；回滚方案不变。
  - 剩余风险：同 G1-03（补定的结构可能需要调整）。
  - 下一步：FIX-1（优先于 G1-04）。

### [x] G1-04 Envelope 校验、版本协商与兼容 fixture
- 依据：PLAN T-01 第 2、6 步；WIRE §1/§2/§3.1；V-02。
- 内容：纯函数校验放在 `contracts/validate`（运行时帧编解码属于 `internal/wire`，不放这里）：按帧种类检查必需字段、拒绝 UNSPECIFIED、拒绝未知 kind、RESPONSE 允许同时带 payload 与 error、ErrorInfo 的 code 必须非零、未知错误码映射为 INTERNAL）；`Negotiate(local, peer)`（Major 必须相等，Minor 取较小值，required 能力必须属于交集）。`contracts/testdata/` 存放二进制 golden fixture 和生成它的测试（加 `-update` 标志）。
- 完成条件：表驱动测试覆盖 V-02 的全部分项（Major 不一致、Minor 取交集、未知必需能力、UNSPECIFIED、未知 kind、payload+error 同时存在、字段存在性、JSON 与 Protobuf 编码区分）；golden 往返测试字节完全一致。
- 回写：
  - 实际做了什么：新增 `contracts/validate`。`envelope.go`：常量 ProtocolMajor=1/Minor=0/SchemaVersion=1；`Envelope()` 按帧种类校验 kind（拒绝 UNSPECIFIED 和未知值）、major、session_id（HELLO 必须为空，其余必须 16 字节）、trace_id 长度、call_id 与 initiator_side（REQUEST/RESPONSE/CANCEL 必须非零且合法，其余帧 call_id 必须为 0）、parent_call_id（只允许插件发起的 REQUEST 设置，且不能显式为 0）、REQUEST 必须有 deadline、error 只能出现在 RESPONSE/SESSION_ERROR 且 code 不能为 0、载荷规则（以 schema_id 判断是否有载荷；控制帧必须用 PROTOBUF 并匹配固定 schema_id；JSON 载荷必须语法合法，否则返回 INVALID_ARGUMENT；RESPONSE 必须有结果或 error）；`DecodeControl()` 在校验之后解码控制帧载荷；`SchemaID()` 生成 `alonabot.<snake>.v1`；`PublicCode()` 把未知错误码映射为 INTERNAL。`negotiate.go`：`Negotiate(local, peer)` 要求 Major 相等，Minor 取较小值，能力按 name+version 精确求交集，双方的 required 都必须在交集中，失败返回 FAILED_PRECONDITION。golden fixture 共 4 个：`contracts/testdata/frames/{hello,request_with_parent,response_partial_with_error,session_error}.bin`（使用 Deterministic marshal，`go test ./contracts/validate/ -update` 重新生成）。WIRE §5A 补充了 schema_id 约定。
  - V-02（T-01 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 ./...` → contracts 与 contracts/validate 都 ok，共 39 个子测试通过，覆盖：Major 不一致（Envelope 与 Negotiate 两处）、Minor 取较小值（两个方向）、未知必需能力、能力版本不同不算匹配、kind 为 UNSPECIFIED 或未知值、initiator_side 为 UNSPECIFIED 或未知值、parent_call_id 显式为 0（字段存在性）、payload 与 error 同时存在（部分发送）、JSON 与 Protobuf 编码区分（控制帧使用 JSON 被拒；JSON 编码却装着 Protobuf 字节返回 INVALID_ARGUMENT）。golden：再次执行 `-update` 后 `git diff` 为空（稳定）；把 fixture 的 target 改掉之后 `TestGoldenFrames/request_with_parent.bin` 失败（"编码与 golden 不一致"），说明能检测到编码漂移。
  - 剩余风险：golden 依赖 protobuf-go 的 Deterministic 编码，升级 protobuf 版本时可能需要用 `-update` 重新生成，并人工确认差异；V-02 中的「无注册残留」属于 REGISTER 原子性，交给 G1-16（T-02）和 T-04；依赖 Session 状态的检查（source_plugin_id 与认证 ID 一致、父调用是否在途）交给 G1-14/G1-16。
  - 下一步：G1-05。

### [x] G1-05 manifest、插件 ID 与外部 release 清单
- 依据：CONTRACTS §9.1/§9.2、§7.1；V-01、V-41（T-01 分项）。
- 内容：`contracts/manifest`：解析 plugin.yaml（go.yaml.in/yaml/v3，禁止未知字段），manifest_version=1，市场/本地 ID 正则，长度 64，Windows 保留设备名，统一小写，SemVer，protocol_min/max_major，os/arch 限定为五个目标组合，相对 entrypoint 不得包含 `..` 或绝对路径，migration_mode/prepare_effects 枚举；本地 ID 缺市场前缀只给警告。`contracts/release`：release.json 结构，Ed25519 对文件原始字节验签（先验签后解析），资产 URL/长度/SHA256 字段。fixture 放在 `contracts/testdata/manifest/`、`release/`。
- 完成条件：测试覆盖：合法市场 ID、本地 ID 产生警告、路径穿越被拒绝、缺失 ID、保留名、大写、不支持的 os/arch、未知 migration_mode、签名正确、签名错误、重新序列化后验签失败（证明验的是原始字节）。
- 回写：
  - 实际做了什么：引入已批准的依赖 `go.yaml.in/yaml/v3 v3.0.5`。`contracts/manifest`：`Platforms`/`SupportedPlatform`（五个目标组合）；`ValidateID`（市场正则、64 字符上限、带市场前缀但格式错误、大写字母、本地安全 slug、Windows 保留设备名，本地 ID 返回缺前缀警告）；`ValidVersion`（SemVer 2.0 官方正则）；`Parse`（`KnownFields(true)` 拒绝未知字段，只允许一个 YAML 文档；必填项用指针区分缺失与零值；manifest_version=1、协议主版本范围、required_capabilities 必须出现（允许空列表）、os/arch、entrypoint、config_version、source；migration_mode/prepare_effects 只接受规定的枚举值，未知值报错，未声明时给出「不允许自动升级」警告）；`ValidateRelPath`（必须是 / 分隔的规范相对路径，不允许 .. 或 .、反斜杠、盘符、NUL）。`contracts/release`：`KeyID`（`ed25519:` + 公钥 SHA256 十六进制）；`DecodeSignature`（sig 文件为标准 base64 编码的 64 字节，允许末尾换行）；`Verify` **先对原始字节执行 ed25519.Verify，再严格解析**（DisallowUnknownFields，对象之后不得有多余内容），然后核对 publisher_key_id 与验签公钥一致，以及插件 ID、SemVer、资产（平台属于五个目标且不重复、size>0、sha256 为 64 位小写十六进制、https 绝对 URL）。fixture：`contracts/testdata/manifest/{market,local_no_prefix}.yaml`、`contracts/testdata/release/release.json`（公钥标识是占位值，测试时替换成固定测试密钥的 KeyID 后再签名）。
  - V-01 / V-41（T-01 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 ./...` 中 4 个包都 ok；manifest 与 release 共执行 36 个用例并全部 PASS，覆盖：合法市场 ID、本地 ID 只给警告、路径穿越 ID 与 entrypoint、缺失 ID、Windows 保留名、大写 ID、不支持的 os/arch、未知 migration_mode/prepare_effects、未知字段、多文档；签名正确、内容改动后验签失败、非信任公钥、损坏的 sig 文件；**重新序列化后的 JSON 无法通过原签名**（证明验签对象是原始字节）；签名有效但内容非法（非 https、大写摘要、重复平台、非法 ID/版本、公钥标识不符）也会被拒绝。每个变异用例都先断言确实改动了 fixture，避免出现空转的测试。`gofmt -l contracts` 为空。
  - 剩余风险：`source` 只检查非空，格式（例如 github.com/owner/repo）以及与已绑定来源是否一致留给 T-09；V-01 中「重复 ID 的先后顺序稳定 / first-wins」属于安装记录，交给 T-09；ID 大小写冲突检测（不能靠大小写形成两个安装）在 manifest 层面已经通过拒绝大写来保证，安装层仍需按规范小写比较。
  - 下一步：G1-06。

### [ ] G1-06 MessageChain 结构校验
- 依据：MESSAGE-MODEL §2；CONTRACTS §8；PLAN T-01 第 3 步（「为首发消息建立实际 Schema 与验证入口」）。
- 内容：`contracts/validate` 中加入纯函数 `ValidateChain`：媒体来源三选一、reference/buttons 每条最多一个、只含元数据的 Chain 被拒绝、空 Chain 被拒绝、text 的 format 与 entities 互斥、entity 的 UTF-16 偏移落在完整字符边界且不越界、按钮 action_id 非空、filename 不含路径。只做结构检查，平台能力检查属于 T-03。
- 完成条件：每条规则至少有一个合法用例和一个非法用例；错误统一返回 INVALID_ARGUMENT 并说明是哪个 segment 的问题。
- 回写：

### [ ] REVIEW-2 集中检查（G1-04～G1-06）
- 检查清单同 REVIEW-1；另外核对校验函数是否只依赖 contracts，以及 fixture 能否重复生成。
- 回写：

### [ ] G1-07 SDK 最小公共类型与可编译示例
- 依据：PLAN T-01 第 4 步；DESIGN §D/§J；CONTRACTS §4.1/§5.2/§6.2；plan.md 假设 3。
- 内容：`sdk`：`plugin.Base`（或 `sdk.Base`，以 DESIGN 示例的可编译形式为准）、`Context`（StopPropagation、Send 签名）、两类 Handler 的类型约束、CommandRegistry/Group 的显式注册、OnStart/OnReady/OnStop 接口、HTTP Handler 注册、调度声明；`sdk/message`（Chain、Text 等构造函数）；`sdk/config`（ConfigLoader/ConfigValidator）；`sdk/gormdb`（SelfMigrator、Migration、VersionedMigrator 接口，以及互斥检查函数）。`examples/plugins/minimal` 包含一个根插件、两类 Handler、StopPropagation、Config 的 OnLoad/Validate、HTTP Handler 和一个调度声明。
- 完成条件：`go build ./...` 通过；有测试证明同时实现 Migrate 和 Migrations 会被互斥检查拒绝；没有运行行为的入口返回明确错误，不 panic。
- 回写：

### [ ] G1-08 SQLite/GORM 组合验证
- 依据：PLAN T-01 第 5 步；CONTRACTS §6.1/§6.2；V-26、V-39（T-01 分项）。
- 内容：引入 gorm、libtnb/sqlite、modernc.org/sqlite。在 `sdk/gormdb` 中实现打开受管 SQLite 的函数：单连接，每个物理连接都设置 WAL、synchronous=FULL、foreign_keys=ON、busy_timeout=5000，并回读核对；测试覆盖 CRUD、Tx 提交与回滚、通过 modernc Backup API 把运行中的 WAL 库备份到候选文件并从副本读出、Up 与 `_alonabot_migrations` 账本在同一事务内提交（Up 失败时账本也没有记录）。
- 完成条件：`go test -race ./sdk/gormdb/...` 通过；实际 PRAGMA 回读值写进测试断言；go.mod 锁定通过验证的版本。
- 回写：

### [ ] G1-09 PostgreSQL 分支验证
- 依据：PLAN T-01 第 5 步；CONTRACTS §6.3；V-26。
- 内容：引入 gorm.io/driver/postgres。测试通过 `ALONABOT_TEST_PG_DSN` 启用：以超级用户 DSN 创建专属 role/schema（名称经过标识符清洗，不直接使用插件 ID），再用插件 role 连接，核对 search_path 只包含该 schema，执行 CRUD；事务失败后数据保持不变；插件 role 不能访问其他 schema。按 docs/development.md 的命令启动临时容器跑一次，然后删除容器。
- 完成条件：未设置环境变量时测试 Skip；在容器上实际运行通过并记录命令与输出；容器已清理（`docker ps -a` 中没有 alonabot-test-pg）。
- 回写：

### [ ] REVIEW-3 集中检查（G1-07～G1-09）
- 检查清单同上；另外核对依赖都在已批准清单内（`go list -m all` 中的直接依赖）、`go mod tidy` 后没有差异，并用 code-review-expert 审查到目前为止的分支 diff。
- 回写：

### [ ] G1-10 架构 import 守卫
- 依据：ARCHITECTURE.md「依赖方向」；docs/development.md 规则表。
- 内容：新增测试（例如 `internal/archtest`），调用 `go list -json ./...` 读取每个包的 Imports，按 ARCHITECTURE 的允许边逐条检查，未列出的模块内边视为失败，错误信息包含源包、目标包和违反的规则。负例：testdata 中放一个小型 fixture module，包含一条禁止边（例如 contracts → sdk），断言守卫会报告它。
- 完成条件：当前代码通过；负例确实被拒绝；docs/development.md 规则表的状态改为「已实现 / 已观测」。
- 回写：

### [ ] G1-11 五平台交叉编译与 T-01 收尾
- 依据：PLAN T-01 第 7 步与「可审查交付」；V-01/V-02/V-26/V-32/V-37/V-39/V-41 的 T-01 分项。
- 内容：执行 plan.md §6 中的五平台命令，记录每个组合的结果；整理 T-01 交付清单与证据表（写进本 task 的回写区，并在 docs/development.md 中更新版本与规则表）；确认 T-01 的停止条件全部满足。
- 完成条件：五个组合全部构建成功（失败则定位修复或标为阻塞）；T-01 证据表完整，未运行的项写明「未运行」。
- 回写：

## T-02 Core 状态、Native 进程、双 IPC 与基础生命周期

### [ ] G1-12 Core 状态库、core.lock 与提交原语
- 依据：PLAN T-02 第 1 步；CONTRACTS §6.1、§12.2；V-36（T-02 分项）。
- 内容：`internal/core/state`：用 x/sys 的 flock/LockFileEx 获取 `<data_dir>/core.lock`；打开 state.sqlite（modernc，单连接，设置并核对 PRAGMA）；内建版本号 schema 迁移；最小表结构（plan.md 假设 6）；在一个事务内提交 ActiveGeneration+COMMITTED，并核对预期的旧代次；按插件串行化管理操作。状态库损坏或打不开时报错退出，不能重建空库。
- 完成条件：测试覆盖：同一 data_dir 启动第二个 Core 被拒绝（用子进程或两个独立句柄验证）、预期旧代次不一致时提交失败、损坏的状态库不会被重建、PRAGMA 回读核对、同插件并发操作被串行化。
- 回写：

### [ ] G1-13 帧编解码与 TCP/stdio 传输
- 依据：PLAN T-02 第 3 步；CONTRACTS §1.1；WIRE §1；V-03。
- 内容：`internal/wire`：完整读取长度前缀和载荷，先校验长度再分配缓冲（1～4,194,304 字节），零长度、超限、截断、无法解码都返回明确错误；每个连接只有一个串行 writer；TCP 用 `net.Listen("tcp4","127.0.0.1:0")` 并持有 Listener；stdio 传输（stdout 只传帧，出现非协议字节就让会话失败）。
- 完成条件：同一份 fixture 在 TCP 和 stdio（用 `io.Pipe`/`os.Pipe` 模拟）上的往返结果一致；边界用例（0、4 MiB、4 MiB+1、截断、分多次 Read 才凑齐一帧）全部覆盖；并发写入的帧不交错（`-race`）。
- 回写：

### [ ] REVIEW-4 集中检查（G1-10～G1-13）
- 检查清单同上；另外核对 T-01 停止条件、锁实现的平台 build tag，以及交叉编译是否仍然通过。
- 回写：

### [ ] G1-14 Session 调用引擎：配对、取消、准入与优先通道
- 依据：PLAN T-02 第 4 步；CONTRACTS §1.2；WIRE §4；V-05、V-42（T-02 分项）。
- 内容：`internal/wire` 中实现调用表，以 `(initiator_side, call_id)` 为键（每个 Session 一个实例）；call_id 单调递增；默认 deadline 30s，子请求 deadline 不晚于父请求；parent_call_id 只接受同 Session 内仍在途的 Core→Plugin 请求；准入预算为每个方向 256 个请求、每个方向 16 MiB 字节，另有一个跨 Session 共享的 64 MiB 全局预算对象，满额时立即返回 RESOURCE_EXHAUSTED；Response/Cancel/控制帧走有界优先通道；尚未写出的 Request 被取消时从队列撤出，已写出的 Request 才能随后发送 Cancel；迟到或重复的 Response 只计数并丢弃；payload 与 error 同时保留；控制通道也无法推进时关闭 Session，未完成的调用以 UNAVAILABLE 结束。
- 完成条件：两个方向使用相同 call_id 时互不串号；乱序 Response；业务槽满额时 Response/Cancel 仍能推进；发送前取消的请求从未出现在线上；所有测试在 `-race` 下通过。
- 回写：

### [ ] G1-15 Native 进程启动、runtime.lock 与进程树
- 依据：PLAN T-02 第 2 步；CONTRACTS §7.1、§12.2；V-06、V-36（分项）。
- 内容：`internal/core/process`：根据安装记录取得入口程序，生成 256-bit Session Secret，构造精简环境（不传 Core 管理凭证），Unix 下使用 Setpgid 并对进程组发信号，Windows 下用 Job Object（build tag）；收集 stdout/stderr 日志并标记 plugin_id；取得退出状态。`sdk` 在任何用户回调之前获取 `<plugins>/<id>/runtime.lock`，锁文件不删除；Core 可以在旧进程退出后临时持有该锁，再释放。
- 完成条件：测试用重入的辅助进程验证：日志带插件 ID、环境中没有管理凭证、杀进程组后子进程也被回收（darwin 实测）、runtime.lock 被占用时第二个 SDK 实例启动失败、Windows 代码交叉编译通过（运行状态标为「未运行」）。
- 回写：

### [ ] G1-16 握手、Bootstrap、REGISTER 与 PREPARED
- 依据：PLAN T-02 第 5 步；CONTRACTS §1.1、§4.2；WIRE §3；V-04、V-37（分项）。
- 内容：Core 侧校验 Hello（Secret 恒时比较、plugin_id、build_digest、协议协商）后发出 HelloAck（含 Bootstrap：generation_id、原始配置 JSON（只检查是否为 JSON 对象及大小）、DB 描述、阶段预算）；SDK 侧按顺序执行配置准备（B1 中只有原始 JSON 透传和可选回调的占位）→ OnStart → 发送 REGISTER；Core 原子校验整份注册集（重复 route ID、超出 4 MiB 时返回 RESOURCE_EXHAUSTED），失败时不留下注册残留；SESSION_ERROR 携带失败阶段；握手 10s、OnStart 60s 的阶段 watchdog（测试中注入缩短的预算）；Bootstrap 内容不写入日志。
- 完成条件：测试覆盖：伪造 Secret、跨插件 Secret、协议 Major 不一致、未知必需能力、REGISTER 超限、重复 route、OnStart 返回错误、OnStart 挂起超过预算；日志中找不到 Secret/Bootstrap 内容。
- 回写：

### [ ] REVIEW-5 集中检查（G1-14～G1-16）
- 检查清单同上；重点检查并发与竞态（`-race` 跑 10 次：`go test -race -count=10` 覆盖 wire 和 session 相关包）、错误路径上的资源泄漏（goroutine、fd、Listener），并用 code-review-expert 审查。
- 回写：

### [ ] G1-17 提交与激活门禁
- 依据：PLAN T-02 第 6 步；CONTRACTS §4.1、§12.1；V-21、V-37（分项）。
- 内容：PREPARED 之后，Core 在状态库事务中提交 ActiveGeneration+COMMITTED（事务内不等待 RPC），然后发送 ACTIVATE；SDK 执行 OnReady（30s 预算），由唯一 writer 写出 ACTIVATED 后打开 ready gate；Core reader 在应用 Activated 状态之后才处理后续 REQUEST；Preparing/Prepared/Activating 阶段收到业务 REQUEST 一律拒绝；提交后 OnReady 失败进入 Degraded，不回到旧代次；普通 restart 复用原有提交记录和 Generation。
- 完成条件：测试覆盖：激活前的业务 REQUEST 被拒绝；ACTIVATED 与紧随其后的 REQUEST 在同一连接上按顺序处理；OnReady 失败后状态为 Degraded 且状态库中仍是新代次；restart 不产生新 Generation。
- 回写：

### [ ] G1-18 停止、阶段超时与有限恢复
- 依据：PLAN T-02 第 7 步；CONTRACTS §4.2；V-06、V-21（分项）。
- 内容：停用时冻结新的根调用，最多排空 6s（父链内合法的子调用仍可完成）→ Cancel 剩余调用 → 发送 STOPPING，deadline 为 `min(t0+10s, 发送时刻+4s)` → SDK 用独立 context 执行 OnStop → 到 t0+10s 时终止进程树；Core Shutdown 并行停止所有插件，共享同一个截止时间；EOF 时旧 Session 立即失效，SDK 收到 EOF 后退出；5 分钟内第三次异常退出时自动禁用，正常退出不重启，手动启用会重置失败窗口；用户函数不响应 context 时由 watchdog 结束进程。
- 完成条件：测试（注入缩短的时间常量，另有测试核对生产常量等于契约数值）覆盖：drain 期间子调用完成、STOPPING deadline 公式、OnStop 挂起时进程在总截止时间被杀、多个插件并行停止的总耗时不超过一份预算、崩溃计数与自动禁用、正常退出不重启。
- 回写：

### [ ] G1-19 cmd/alonabot 装配与真实进程验收切片
- 依据：PLAN T-02 第 8 步与「对照现有验收」。
- 内容：`cmd/alonabot` 提供开发用的启停入口；`examples/plugins/minimal` 以真实子进程分别在 TCP 和 stdio 上跑完整生命周期；故障辅助进程覆盖挂起、错帧、stdout 日志污染、迟到响应、突然退出；场景包括双 Core、同插件并发操作、Core 被 kill 后 SDK 断链退出；检查进程和 Listener 都已关闭。把实际命令写入 docs/development.md。
- 完成条件：所有场景测试通过并可以重复运行（`-count=3`）；整理 V-03/V-04/V-05/V-06/V-21/V-36/V-37/V-42 的 T-02 分项证据表，明确哪些留待 T-07/T-08/T-09/T-10 接入；Linux/Windows 的运行状态记为「未运行」。
- 回写：

### [ ] REVIEW-6 集中检查（G1-17～G1-19）
- 检查清单同上；另外核对 T-02「停止和失败条件」逐条满足，并用 code-review-expert 审查整个分支。
- 回写：

---

## 终审（全部 task 完成后）

- 对照 input.md 与 plan.md §1 检查完成定义；最大范围复查（代码质量、并发、安全、错误处理、测试覆盖、交叉编译、文档、回滚方案）；汇总所有阻塞项与「未运行」项；在本节写出结论。
- 结论：

## 追加的修复 task

（复查中发现的问题追加在这里）

### [x] FIX-1 补 ServiceInvocation 载荷并同步文档（来自 REVIEW-1，排在 G1-04 之前）
- 依据：WIRE §4（可信上下文随 Core→Plugin Invocation 载荷下发）、CONTRACTS §3。
- 内容：在 `invocation.proto` 中新增 `ServiceInvocation{route_id, context InvocationContext, encoding PayloadEncoding, schema_id, schema_version, payload}`，作为 Core→目标插件的服务调用载荷（服务响应仍直接把业务载荷放在 Envelope 中，使用 output schema）；在 WIRE §5A 中补充说明；更新 ARCHITECTURE.md 的「当前状态」和 contracts 行（含 `contracts/internal/protocgen` 生成工具）；把 docs/development.md 规则表中字段号测试一行的状态改为「已观测」。
- 完成条件：生成、build、vet、test 通过，生成稳定；测试断言 ServiceInvocation 包含 context 和 payload 字段。
- 回写：
  - 实际做了什么：`invocation.proto` 新增 `ServiceInvocation`（route_id=1、context=2 InvocationContext、encoding=3、schema_id=4、schema_version=5、payload=6），为此新增对 envelope.proto 的 import；`wirespec_test.go` 新增 `TestServiceInvocationCarriesContext`；WIRE §5A 补充这条说明；ARCHITECTURE.md 的当前状态与 contracts 行改为「部分实现（proto 与生成）」，并写明 protocgen 工具；docs/development.md 中字段号测试一行改为「已观测」，并写明测试位置。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go generate ./... && go vet ./... && go build ./... && go test -race -count=1 ./contracts/` → ok；stage 之后再执行 `go generate ./...`，`git diff --stat` 为空。
  - 剩余风险：ServiceInvocation 的使用方（T-06 Broker）尚未实现，字段是否足够要到 T-06 才能验证。
  - 下一步：G1-04。
