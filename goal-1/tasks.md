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
  - 实际做了什么：`go get -tool` 固定 protoc-gen-go v1.36.12（go.mod 中 `tool` 指令 + `require google.golang.org/protobuf v1.36.12`）；新增 `contracts/doc.go`（`//go:generate go run ./internal/protocgen`）和 `contracts/internal/protocgen/main.go`（核对 `protoc --version` == `libprotoc 36.0` → 删除旧的 `pb/*.pb.go` → 用 `go tool -n protoc-gen-go` 得到的插件路径执行 protoc，参数为 `--go_opt=module=github.com/AlonaBot/AlonaBot/contracts`）。**与原计划不同：** 没有写占位的 version.proto，而是直接用 WIRE §2 已完整定义的 `errors.proto`（ErrorCode 0～17、ErrorInfo 1～3）验证整条流水线，避免留下无用文件；G1-02 不再需要写 errors.proto，只需补它的字段号测试。docs/development.md 的环境表已写入实际版本。
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
  - 验证证据（darwin/arm64，go1.27.1，libprotoc 36.0，本 task 的 commit）：`go generate ./... && go build ./... && go vet ./... && go test -race ./...` → `ok github.com/AlonaBot/AlonaBot/contracts`；stage 之后再执行 `go generate ./...`，`git diff --stat` 为空。变异检查：把 trace_id 的字段号临时改成 17 → `TestEnvelopeFieldNumbers` 失败（"Envelope 字段 11 应为 trace_id，实际 <nil>"），还原后通过。
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

### [x] G1-06 MessageChain 结构校验
- 依据：MESSAGE-MODEL §2；CONTRACTS §8；PLAN T-01 第 3 步（「为首发消息建立实际 Schema 与验证入口」）。
- 内容：`contracts/validate` 中加入纯函数 `ValidateChain`：媒体来源三选一、reference/buttons 每条最多一个、只含元数据的 Chain 被拒绝、空 Chain 被拒绝、text 的 format 与 entities 互斥、entity 的 UTF-16 偏移落在完整字符边界且不越界、按钮 action_id 非空、filename 不含路径。只做结构检查，平台能力检查属于 T-03。
- 完成条件：每条规则至少有一个合法用例和一个非法用例；错误统一返回 INVALID_ARGUMENT 并说明是哪个 segment 的问题。
- 回写：
  - 实际做了什么：新增 `contracts/validate/chain.go`，入口命名为 `validate.Chain`（与 `validate.Envelope` 保持一致，而不是任务描述里的 ValidateChain）。检查项包括：空 Chain 或 nil；片段未设置类型；reference 和 buttons 各自最多一个；只有元数据、没有可发送内容（text/image/video/file/mention/forward 才算内容）；text 不能为空且必须是合法 UTF-8；format 显式设置时不能是 UNSPECIFIED 或未知值，并且与 entities 互斥；entity 的类型必须合法、长度大于 0、不越界（以 UTF-16 码元计），起止都要落在字符边界上（不能截断代理对）；url 只能且必须用于 text_link（必须是 http(s) 绝对地址），external_user_id 只能且必须用于 text_mention，language 只能用于 pre；媒体来源恰好一个，media_ref/incoming token 不能为空，https_url 必须是 https 绝对地址；filename 不能包含路径分隔符、NUL 或 ./..；mention 必须有用户；reference 必须有消息 ID，带 target 时字段要齐全；buttons 至少一行，每行至少一个按钮，action_id 与 label 不能为空，value 必须是合法 JSON；forward 至少一个条目，嵌套 Chain 递归校验，报错带完整位置（例如 `segment[0].forward.items[1].segment[0]`）。错误码统一为 INVALID_ARGUMENT。**由我补定的规则：** 空文本片段视为非法；entity 禁止携带与类型不符的附加字段；按钮 value 按 JSON 文本校验（与 G1-03 的约定一致）。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 ./...` 中 4 个包都 ok；TestChain 的子测试加上 TestChainErrorNamesSegment 共 43 个 PASS（8 个合法用例、34 个非法用例，每条规则至少各有一个），每个非法用例都同时断言错误文本和 INVALID_ARGUMENT；`gofmt -l contracts` 为空（测试文件格式化过一次）。
  - 剩余风险：平台相关的限制（文本或 caption 长度、markdown 方言、相册数量、按钮数量上限、mention 能否渲染）属于 T-03 的发送计划；value 是否符合 action Schema、media_ref 是否有效以及租约、Target 授权，都属于 Core 运行时检查。
  - 下一步：REVIEW-2。

### [x] REVIEW-2 集中检查（G1-04～G1-06）
- 检查清单同 REVIEW-1；另外核对校验函数是否只依赖 contracts，以及 fixture 能否重复生成。
- 回写：
  - 检查过程与结论（commit 853ea10 之上）：
    1. 是否偏离 input.md：没有偏离；新增依赖只有已批准的 go.yaml.in/yaml/v3。
    2. 依赖边界：`go list` 显示 validate 只 import contracts/pb、protobuf 和标准库；manifest 只 import yaml 和标准库；release 只 import contracts/manifest 和标准库，都符合 ARCHITECTURE（contracts 不 import 本 module 中 contracts 以外的包）。
    3. fixture 可重复生成：`go test ./contracts/validate/ -update` 之后 `git status` 为空。
    4. `go vet ./...`、`gofmt -l contracts`（为空）、`go test -race -count=1 ./...` 的 4 个包都通过；只剩 protocgen 中有意保留的 Fprintln 错误输出，没有其他残留。
    5. 逐个检查行为：Envelope/Negotiate/manifest/release 的边界情况（空 YAML、重复 key、数字形式的 ID、JSON 尾随内容、大写的市场前缀）都按预期拒绝。**发现问题：** `validate.Chain` 为 entity 边界检查建立了 `map[int]bool`，包含文本里每一个字符的边界；用临时探针测试（已删除）测得 4 MiB 文本加 1 个 entity 时 TotalAlloc 增加 **296 MiB**，超过 Core 64 MiB 的 payload 总预算，恶意或有缺陷的插件可以借此放大 Core 内存。→ 登记 FIX-2。
    6. 文档同步：ARCHITECTURE.md 中 contracts 的状态仍是「部分实现（proto 与生成）」，没有写上 validate/manifest/release → 放进 FIX-2。
    7. 安全：release 先验签再解析、manifest 拒绝路径穿越、filename 拒绝路径，都已有测试；没有发现泄漏敏感信息的错误信息（错误只回显 ID 或路径本身，不涉及 secret）。
  - 剩余风险：Chain 的内存问题修复前不能把它接入 Core 的发送路径（最早在 T-03 才会接入，不会影响 T-02）。
  - 下一步：FIX-2（优先于 G1-07）。

### [x] G1-07 SDK 最小公共类型与可编译示例
- 依据：PLAN T-01 第 4 步；DESIGN §D/§J；CONTRACTS §4.1/§5.2/§6.2；plan.md 假设 3。
- 内容：`sdk`：`plugin.Base`（或 `sdk.Base`，以 DESIGN 示例的可编译形式为准）、`Context`（StopPropagation、Send 签名）、两类 Handler 的类型约束、CommandRegistry/Group 的显式注册、OnStart/OnReady/OnStop 接口、HTTP Handler 注册、调度声明；`sdk/message`（Chain、Text 等构造函数）；`sdk/config`（ConfigLoader/ConfigValidator）；`sdk/gormdb`（SelfMigrator、Migration、VersionedMigrator 接口，以及互斥检查函数）。`examples/plugins/minimal` 包含一个根插件、两类 Handler、StopPropagation、Config 的 OnLoad/Validate、HTTP Handler 和一个调度声明。
- 完成条件：`go build ./...` 通过；有测试证明同时实现 Migrate 和 Migrations 会被互斥检查拒绝；没有运行行为的入口返回明确错误，不 panic。
- 回写：
  - 实际做了什么：引入已批准的 `gorm.io/gorm v1.31.2`（它带进来的间接依赖为 jinzhu/inflection、jinzhu/now、golang.org/x/text）。新增 `sdk/plugin`（Base 加 `Root` 接口；Root 带一个未导出方法，只有嵌入 Base 的类型才能满足）、`sdk/message`（Segment 接口、`Chain []Segment`、`Text()`、`Chain.Proto()`）、`sdk/config`（Provider、Loader.OnLoad、Validator，签名与 CONTRACTS §5.2 一致）、`sdk/gormdb`（SelfMigrator、Migration、VersionedMigrator、`CheckMigrators`：两种接口互斥，ID 非零、递增，Up 不能为空）、`sdk`（Error/ErrNotImplemented；Starter/Readier/Stopper；Command/HTTP/ScheduleProvider；Context 提供 Context()、锁存式 StopPropagation、`Send(chain, To(...))`；`Run`；`Collect` 收集注册项并用反射检查两种 Handler 签名——第一个参数必须是 *Context，参数类型只能是 string/int64/float64/bool，返回 error 或 (message.Chain, error)——同时检查同一插件内的重复项：命令路径、HTTP 的 method+path、job_id，以及 HTTP 路径必须以 / 开头、公开路由必须声明验证方式、间隔至少 1s、cron 必须五字段；多个错误合并在一次报告里）。`examples/plugins/minimal`：根插件 Demo，含一个自动回复 Handler（带 StopPropagation）、一个主动发送 Handler（参数为 string+int64）、实现 OnLoad/Validate 的 Config、HTTP GET /status、Every 定时任务，以及 OnStart。Run/Send 在 T-02 前返回 UNSUPPORTED。ARCHITECTURE.md 中 sdk/examples 的状态已同步。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go build ./... && go vet ./... && go test -race -count=1 ./...` 中 7 个包都 ok。`sdk` 测试：3 种合法签名、11 种非法签名、10 类非法注册、错误合并、组路径、StopPropagation 锁存、Send 返回 UNSUPPORTED、Chain.Proto。`gormdb` 测试：4 种合法情况，加上 5 种非法情况（**同时实现 Migrate 与 Migrations 时被拒绝**、ID 为 0、重复、递减、缺少 Up）。示例测试：Collect 识别出 2 个命令、1 个 HTTP 路由、1 个任务，Handler 模式与参数类型正确，Config 实现了 Loader 与 Validator，OnLoad 迁移结果正确；`sdk.Run` 返回 UNSUPPORTED。`go run ./examples/plugins/minimal` 输出「ERROR_CODE_UNSUPPORTED: SDK 运行时尚未实现」，exit 1，没有 panic。`gofmt -l` 为空。
  - 剩余风险：Collect 只是 T-01 的组合证明，没有生成 RouteDescriptor/route_id，也没有覆盖 aliases、Regex/Raw、可选参数、Rule、事件 Handler；这些属于 T-04（生成器与注册索引）。`Send` 的返回类型暂时直接使用 `*pb.SendResult`，T-03 可能改为 SDK 自己的类型。Config 通过 `Config() any` 声明，T-08 可能调整。
  - 下一步：G1-08。

### [x] G1-08 SQLite/GORM 组合验证
- 依据：PLAN T-01 第 5 步；CONTRACTS §6.1/§6.2；V-26、V-39（T-01 分项）。
- 内容：引入 gorm、libtnb/sqlite、modernc.org/sqlite。在 `sdk/gormdb` 中实现打开受管 SQLite 的函数：单连接，每个物理连接都设置 WAL、synchronous=FULL、foreign_keys=ON、busy_timeout=5000，并回读核对；测试覆盖 CRUD、Tx 提交与回滚、通过 modernc Backup API 把运行中的 WAL 库备份到候选文件并从副本读出、Up 与 `_alonabot_migrations` 账本在同一事务内提交（Up 失败时账本也没有记录）。
- 完成条件：`go test -race ./sdk/gormdb/...` 通过；实际 PRAGMA 回读值写进测试断言；go.mod 锁定通过验证的版本。
- 回写：
  - 实际做了什么：引入已批准的 `github.com/libtnb/sqlite v1.2.2` 和 `modernc.org/sqlite v1.61.0`（libtnb 的 go.mod 写的是 modernc v1.56.0，MVS 选择了更高的 v1.61.0；x/sys v0.49.0 作为间接依赖一起进入）。`sdk/gormdb/sqlite.go`：`OpenSQLite(ctx, path)` 在 DSN 中通过 `_pragma` 设置 WAL、synchronous(FULL)、foreign_keys(1)、busy_timeout(5000)（modernc 会对每个新物理连接执行这些设置），连接池固定为 1，打开后用 `ReadPragmas` 回读核对，不一致就关闭并报错；拒绝含 ?/# 的路径；**GORM 日志写到 stderr**（stdio 传输时 stdout 只能传协议帧）。`sdk/gormdb/versioned.go`：`RunVersioned` 先校验迁移列表，再建好账本表 `_alonabot_migrations`，然后对每个尚未完成的 ID 执行一次 `Transaction{Up(tx); 写账本}`，失败时整步回滚并停止执行后续迁移；`AppliedMigrations` 读取账本。备份验证只放在测试里，直接调用 modernc 的 `conn.NewBackup/Step/Finish`（Core 侧的正式实现属于 T-08 的 internal/core/storage）。
  - V-26 / V-39（T-01 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 -v ./sdk/gormdb/` 共 6 个测试 PASS：PRAGMA 回读的实际值为 `{JournalMode:wal Synchronous:2 ForeignKeys:1 BusyTimeout:5000}`，MaxOpenConnections=1；CRUD（Create/Update/First/Delete）；Tx 回滚后数据不可见，提交后可见；外键约束确实生效；**WAL Backup**：确认 `-wal` 文件非空（数据尚未 checkpoint）后备份到候选文件，候选中正好有 3 行，备份之后旧库的写入不出现在候选里，候选的写入也不出现在旧库里；**Up 与账本同事务**：迁移 2 在 DDL 之后返回错误时账本只有 [1]、表 b 不存在（DDL 已回滚），修复后重跑只执行 2，再跑不执行任何迁移，非递增的列表被拒绝。变异检查：去掉 foreign_keys(1) 后 `TestOpenSQLitePragmas` 报「设置未生效 … ForeignKeys:0」，`TestForeignKeysEnforced` 也失败，恢复后通过（去掉 synchronous(FULL) 不会让测试失败，因为 SQLite 默认值本来就是 FULL=2，这一项无法用变异法区分）。`CGO_ENABLED=0 go test ./sdk/gormdb/` ok（纯 Go）。`go mod tidy` 第一次执行时 goproxy.cn 出现一次瞬时的 "protocol error: received DATA after END_STREAM"，重跑 exit 0，`go mod verify` 显示 all modules verified；全仓库 `go test -race ./...` 的 7 个包都 ok。
  - 剩余风险：synchronous=FULL 依赖显式 DSN 设置加回读核对，变异法无法证明 DSN 设置本身在起作用（默认值相同）；回读核对只在打开时对第一个物理连接执行，后续重连依赖驱动按 DSN 重新设置（单连接池且 ConnMaxLifetime 为 0，正常情况下不会重连）；在事务里误用外层 db 会让单连接池自锁，这一点只在文档注释中说明，没有运行时防护。
  - 下一步：G1-09（PostgreSQL）。

### [x] G1-09 PostgreSQL 分支验证
- 依据：PLAN T-01 第 5 步；CONTRACTS §6.3；V-26。
- 内容：引入 gorm.io/driver/postgres。测试通过 `ALONABOT_TEST_PG_DSN` 启用：以超级用户 DSN 创建专属 role/schema（名称经过标识符清洗，不直接使用插件 ID），再用插件 role 连接，核对 search_path 只包含该 schema，执行 CRUD；事务失败后数据保持不变；插件 role 不能访问其他 schema。按 docs/development.md 的命令启动临时容器跑一次，然后删除容器。
- 完成条件：未设置环境变量时测试 Skip；在容器上实际运行通过并记录命令与输出；容器已清理（`docker ps -a` 中没有 alonabot-test-pg）。
- 回写：
  - 实际做了什么：引入已批准的 `gorm.io/driver/postgres v1.6.3`（依赖 pgx/v5 v5.10.0、puddle、pgpassfile、pgservicefile）。`contracts/manifest.PostgresIdentifier(prefix, id)` 生成规则为 prefix + 清洗后的 ID + "_" + SHA256 前 8 位十六进制，最长 63 字节，结果确定，且 `my-tool` 与 `my_tool` 不会映射成同一个名字。`sdk/gormdb.OpenPostgres(ctx, *pb.PostgresConnection)`：schema 必须是小写标识符，search_path 只包含插件 schema；连接之后核对 current_user 和 `current_schemas(false)` 必须正好是 [schema]，否则关闭连接并报错；连接失败的错误信息不包含 DSN 和密码；GORM 日志写到 stderr。测试 `postgres_test.go` 由 `ALONABOT_TEST_PG_DSN` 启用：用超级用户按管理员预置流程创建专属 role 和 schema（插件 ID 带 '-'，所以必须先清洗），插件连接只使用专属 role，测试结束后 DROP SCHEMA/OWNED/ROLE。docs/development.md 补充了就绪检查与测试命令。
  - V-26（T-01 分项）证据（darwin/arm64，go1.27.1，Docker 29.4.2，镜像 postgres:17 a63c1ea3f6a3，本 task 的 commit）：未设置环境变量时，两个测试都 SKIP，并输出「PostgreSQL 测试未运行」。容器命令：`docker run -d --rm --name alonabot-test-pg -e POSTGRES_PASSWORD=alonabot -p 127.0.0.1::5432 postgres:17`（映射到 127.0.0.1:53496），`pg_isready` 返回 accepting connections 之后执行 `ALONABOT_TEST_PG_DSN=postgres://postgres:alonabot@127.0.0.1:53496/postgres?sslmode=disable go test -race -count=1 -v -run Postgres ./sdk/gormdb/`：`TestPostgresPluginRole` PASS（search_path 等于 `ab_test_plugin_ebb10852_775b20fd`；items 表落在插件 schema 中；Create/Update/First 正常；事务回滚后值保持为 b；迁移 2 失败时账本只有 [1] 且 t2 表不存在；读取 `…_other.secret` 返回 `permission denied for schema … (SQLSTATE 42501)`），`TestOpenPostgresRejects` PASS（不存在的 schema 在核对阶段失败；错误密码返回 SASL 认证失败，错误信息中没有密码；非法 schema 名被拒绝）。启用 PG 后全仓库 `go test -race ./...` 的 7 个包都 ok。之后用 psql 查询 `pg_roles` 和 `pg_namespace` 中的 `ab_%`，结果为空（没有残留）。最后 `docker stop alonabot-test-pg`，`docker ps -a` 中残留容器数为 0。
  - 剩余风险：没有在 PG 上测试「首次安装到空 schema、resume 只核对、维护许可」等分支，这些属于 T-08/T-09（V-31/V-39）；超级用户只在测试中用于预置，生产环境由管理员预置；`sslmode=disable` 只用于本机容器；GORM 的 Error 级日志会输出失败的 SQL 文本，日后 SQL 中若带有业务数据，需要在 T-08 评估日志级别。
  - 下一步：REVIEW-3（含 code-review-expert 审查当前分支的 diff）。

### [x] REVIEW-3 集中检查（G1-07～G1-09）
- 检查清单同上；另外核对依赖都在已批准清单内（`go list -m all` 中的直接依赖）、`go mod tidy` 后没有差异，并用 code-review-expert 审查到目前为止的分支 diff。
- 回写：
  - 流程说明：code-review-expert 的 SKILL.md 第 7 步「Next steps confirmation」要求列出审查发现后先询问用户如何处理；goal-mode 的 SKILL.md「红线」禁止在 goal 中向用户提问。按 goal-mode 的规则，审查发现直接登记为修复 task，不暂停等待确认。
  - 审查范围：`main..goal-1`（commit 243b4a3），不含 contracts/pb 生成代码、testdata、goal 文件和 go.sum；共 42 个文件、约 4460 行。contracts 部分已经过 REVIEW-1/2，本次重点是 G1-07～G1-09（sdk、sdk/gormdb、examples）以及跨模块的约束。
  - 机械检查：直接依赖正好是 6 个（libtnb/sqlite v1.2.2、yaml v3.0.5、protobuf v1.36.12、driver/postgres v1.6.3、gorm v1.31.2、modernc v1.61.0），全部在已批准清单内；`go mod tidy` 前后没有差异；`gofmt -l` 为空；`go vet ./...` 和 `go test -race -count=1 ./...` 的 7 个包都 ok。
  - Code Review Summary：COMMENT（没有 P0/P1）。
    - **P2-1 sdk/gormdb/sqlite.go:18 gormLogger 会把绑定参数插进日志**：通过阅读 gorm v1.31.2 的 callbacks.go:142-145 确认，在 `ParameterizedQueries=false`（默认值）时，Trace 用 `Dialector.Explain(sql, vars...)` 把参数值拼进 SQL 文本；Warn 级别会记录慢查询，Error 级别会记录失败的 SQL，业务数据或 PII（插件甚至可能把 secret 写进 DB）就会进入 stderr，再被 Core 采集到普通日志，违反 CONTRACTS §5.3「secret 原文不进入普通日志」。可利用条件：任何慢查询或错误 SQL 都会触发。修复：在 logger.Config 中设置 `ParameterizedQueries: true`，并加一个测试捕获日志输出，断言其中不含参数值。
    - **P2-2 sdk/registry.go:142 HTTP 路径没有做规范化检查**：`/a/../admin`、`//x`、`/a/` 这类非规范路径也能注册，同一资源可能对应多个写法，代理在匹配路由时会产生歧义（CONTRACTS §7.2：Core 按 route 匹配后剥离前缀）。修复：要求 `path.Clean(p) == p`（`/` 本身允许），并拒绝 `..` 段。
    - **P3-1 sdk/gormdb/versioned.go 账本一致性**：没有检测账本中存在但声明里没有的 ID，也没有检测「后面的 ID 已完成、前面的 ID 未完成」这种空洞。CONTRACTS §6.3 的 resume 核对属于 T-08，**推迟**，作为 T-08 的输入记录在这里。
    - **P3-2 ARCHITECTURE.md 中 sdk 行的描述已过时**：gormdb 现在已有 OpenSQLite/OpenPostgres/RunVersioned，描述仍写着「迁移接口与互斥检查」→ 放进 FIX-3 一起同步。
    - **P3-3 sdk.ErrNotImplemented 是可变的共享指针**：调用方可能修改它。影响很小，T-02 实现 Run/Send 后会删除这个错误，**不修**。
  - 其他检查结论：plugin.Root 用未导出方法确保只有嵌入 Base 的类型才能满足它；Collect 的重复检测覆盖了嵌套组；OpenPostgres 的错误不含密码（G1-09 实测）；没有 SQL 拼接注入面（测试中的 DDL 拼接只出现在测试预置代码里，使用的是经过清洗的标识符）；没有发现并发共享状态。
  - 下一步：FIX-3（优先于 G1-10）。

### [x] G1-10 架构 import 守卫
- 依据：ARCHITECTURE.md「依赖方向」；docs/development.md 规则表。
- 内容：新增测试（例如 `internal/archtest`），调用 `go list -json ./...` 读取每个包的 Imports，按 ARCHITECTURE 的允许边逐条检查，未列出的模块内边视为失败，错误信息包含源包、目标包和违反的规则。负例：testdata 中放一个小型 fixture module，包含一条禁止边（例如 contracts → sdk），断言守卫会报告它。
- 完成条件：当前代码通过；负例确实被拒绝；docs/development.md 规则表的状态改为「已实现 / 已观测」。
- 回写：
  - 实际做了什么：新增 `internal/archtest/arch_test.go`（只有测试文件，不进入 `go build ./...`）：在指定目录运行 `go list -json ./...`，按包的 Module.Path 去掉模块前缀，按最长前缀把包归入 ARCHITECTURE 中登记的模块；对 Imports、TestImports、XTestImports 逐条检查：同模块内部的 import 放行，跨模块的边必须出现在 `allowed` 中，不属于任何已登记模块的包报违规；违规信息包含源包、目标包、违反的规则和修复途径。负例 fixture `internal/archtest/testdata/fixture`（独立 module example.com/fixture）：`contracts/bad -> sdk`（禁止）、`sdk -> contracts/ok`（允许）、`tools/x`（未登记）。ARCHITECTURE.md 同步：说明「模块」指的是路径前缀，同模块内部不受限，规则也覆盖测试代码；登记 internal/archtest 一行；Adapter 接口包名固定为 `internal/core/adapter`；检查一节写上守卫的用法与维护方式。docs/development.md 规则表中该项改为「已观测」。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -count=1 -v ./internal/archtest/`：`TestModuleImportsFollowArchitecture` PASS（当前代码没有违规），`TestGuardRejectsForbiddenEdges` PASS（正好报出 2 条违规：`contracts/bad -> sdk：…` 和 `tools/x：不属于…`，合法的 sdk→contracts 没有被误报）。真实仓库中的负例：临时添加 `contracts/validate/zz_bad.go`（import sdk）之后，守卫报出「contracts/validate -> sdk：模块 contracts 不允许 import sdk（见 ARCHITECTURE.md…）」并 FAIL；删除后恢复 PASS，工作区干净。`go list ./... | grep -c fixture` = 0（fixture 不进入主 module）。`go vet ./...` 通过，`gofmt -l internal` 为空。
  - 剩余风险：守卫依赖 PATH 中的 go 命令（与 go test 使用的是同一套工具链）；`allowed` 与 ARCHITECTURE 表格要靠人工保持一致（两处都写明了修改时必须同步）；目前只有测试，尚未接入 CI（仓库还没有 CI）。
  - 下一步：G1-11（五平台交叉编译与 T-01 收尾）。

### [x] G1-11 五平台交叉编译与 T-01 收尾
- 依据：PLAN T-01 第 7 步与「可审查交付」；V-01/V-02/V-26/V-32/V-37/V-39/V-41 的 T-01 分项。
- 内容：执行 plan.md §6 中的五平台命令，记录每个组合的结果；整理 T-01 交付清单与证据表（写进本 task 的回写区，并在 docs/development.md 中更新版本与规则表）；确认 T-01 的停止条件全部满足。
- 完成条件：五个组合全部构建成功（失败则定位修复或标为阻塞）；T-01 证据表完整，未运行的项写明「未运行」。
- 回写：
  - 实际做了什么：按 plan.md §6 执行五平台交叉编译，并补充了构建测试二进制的检查；在 Apple Silicon 上用 `GOARCH=amd64` 通过 Rosetta 实际运行全部测试。运行中发现 `internal/archtest` 内部调用的 `go list` 在该环境下取 VCS 状态失败（"error obtaining VCS status … Use -buildvcs=false"），改为 `go list -buildvcs=false`：import 分析用不到 VCS 信息，修改后在没有 git 或取不到 VCS 状态的环境中也能运行。这是本 task 验证时发现的 G1-10 缺陷，修复只有一行。docs/development.md 补上了交叉编译命令，规则表中这一项改为「已观测（只验证了构建）」。
  - 五平台结果（go1.27.1，CGO_ENABLED=0，本 task 的 commit）：linux/amd64、linux/arm64、windows/amd64、darwin/amd64、darwin/arm64 上的 `go build ./...` 与 `go vet ./...` 全部 OK；linux/amd64、linux/arm64、windows/amd64、darwin/amd64 的 `go test -c ./sdk/gormdb/` 都能编译出测试二进制。运行情况：darwin/arm64 原生 `go test -race ./...` 的 8 个包都 ok；darwin/amd64 经 Rosetta 运行 `GOARCH=amd64 CGO_ENABLED=0 go test ./...`，修复后 8 个包都 ok（生成的测试二进制为 Mach-O x86_64，属于转译运行，不等同于原生 Intel Mac）；linux/*、windows/amd64 **未运行**。
  - **T-01 证据表**（按 TEST-MATRIX §3；commit 以 goal-1 分支为准；平台为 darwin/arm64、go1.27.1，另有说明的除外）：
    | V | T-01 分项内容 | 测试位置 / 命令 | 结果 | 留待 |
    | --- | --- | --- | --- | --- |
    | V-01 | 市场与本地 ID、缺失/路径穿越/保留名/大写、os/arch 不匹配 | `contracts/manifest`（`go test ./contracts/manifest/`） | 通过 | T-04 注册冲突；T-09 安装来源、first-wins |
    | V-02 | Major/Minor、未知必需能力、UNSPECIFIED/未知 kind、字段存在性、payload+error、JSON/Protobuf 编码区分 | `contracts/validate`、`contracts/wirespec_test.go` | 通过 | T-02 握手拒绝；T-04 注册原子性 |
    | V-26 | GORM SQLite/PG 的 CRUD/Tx，Migrate 与 Migrations 互斥 | `sdk/gormdb`（PG 使用 `ALONABOT_TEST_PG_DSN`，postgres:17 容器） | 通过（PG 实测一次，容器已删除） | T-08 SDK 实际连接与凭证边界 |
    | V-32 | 五平台 CGO_ENABLED=0 构建 | 上方交叉编译命令 | 构建 5/5 通过；运行：darwin/arm64 原生通过，darwin/amd64 经 Rosetta 通过，linux/windows 未运行 | T-10 五平台真实运行 |
    | V-37 | Bootstrap/REGISTER 类型与 4 MiB 上限的承载结构 | `control.proto`、`route.proto` 及字段号测试 | 部分：类型已固定；**4 MiB 上限尚未实施**（帧编解码在 G1-13，Bootstrap/REGISTER 超限在 G1-16） | G1-13/G1-16、T-08、T-09 |
    | V-39 | Up 与 `_alonabot_migrations` 同一事务；失败不记录成功 | `sdk/gormdb` 中的 SQLite 与 PG 测试 | 通过（两个后端） | T-08/T-09 首装、resume 与维护许可 |
    | V-41 | 外部 release.json 对确切字节验签，没有自包含摘要循环 | `contracts/release` | 通过 | T-09 真实验签与更新预检 |
  - **T-01 停止条件逐条核对**（PLAN T-01「停止和失败条件」）：契约能编译 ✓（9 个 proto，五平台构建通过）；示例签名成立 ✓（`examples/plugins/minimal` + `sdk.Collect` 测试）；生成稳定 ✓（重复 `go generate` 无 diff，golden fixture 重新生成无 diff）；默认数据库组合的关键事务与备份路径通过 ✓（G1-08）；PG 分支已在容器上实测，没有阻塞 ✓（G1-09）；依赖版本已写入 go.mod/go.sum 并经验证 ✓。**结论：T-01 完成，可以交付给 T-02。**
  - 剩余风险：Rosetta 转译运行只能作为 darwin/amd64 的近似证据；linux/windows 没有任何运行证据；V-37 的上限要在 G1-13/G1-16 才实施；没有 CI，以上检查都是手动执行的。
  - 下一步：G1-12（Core 状态库、core.lock 与提交原语）。

## T-02 Core 状态、Native 进程、双 IPC 与基础生命周期

### [x] G1-12 Core 状态库、core.lock 与提交原语
- 依据：PLAN T-02 第 1 步；CONTRACTS §6.1、§12.2；V-36（T-02 分项）。
- 内容：`internal/core/state`：用 x/sys 的 flock/LockFileEx 获取 `<data_dir>/core.lock`；打开 state.sqlite（modernc，单连接，设置并核对 PRAGMA）；内建版本号 schema 迁移；最小表结构（plan.md 假设 6）；在一个事务内提交 ActiveGeneration+COMMITTED，并核对预期的旧代次；按插件串行化管理操作。状态库损坏或打不开时报错退出，不能重建空库。
- 完成条件：测试覆盖：同一 data_dir 启动第二个 Core 被拒绝（用子进程或两个独立句柄验证）、预期旧代次不一致时提交失败、损坏的状态库不会被重建、PRAGMA 回读核对、同插件并发操作被串行化。
- 回写：
  - 实际做了什么：**新增模块 `internal/oslock`**（ARCHITECTURE 与 archtest 同步更新，新增允许边 sdk→internal/oslock、internal/core→internal/oslock；oslock 本身不 import 本 module 的任何包）。新增理由：core.lock 与 runtime.lock 必须共用同一份锁实现，而 sdk 和 internal/core 之间不能互相 import。`TryLock` 以 0600 打开或创建锁文件，然后非阻塞加独占锁：unix 使用 `unix.Flock(LOCK_EX|LOCK_NB)`（EINTR 时重试，EWOULDBLOCK 视为已被占用，返回 ErrLocked），windows 使用 `LockFileEx(EXCLUSIVE|FAIL_IMMEDIATELY)`（ERROR_LOCK_VIOLATION 返回 ErrLocked）；`Unlock` 重复调用不会出错，锁文件保留不删除。`golang.org/x/sys` 由此变为直接依赖（已批准）。新增 `internal/core/state`：`Open` 先取得 core.lock（被占用时返回 ErrDataDirLocked），再打开 state.sqlite（modernc，单连接，DSN 中设置 WAL/FULL/外键/busy_timeout 后回读核对）；**已有文件先以只读、不带 PRAGMA 的 URI 做预检**（必须存在 schema_version 表，quick_check 必须为 ok），通过后才以可写方式打开，不满足就报错且不改动文件；任何失败都会释放 core.lock；内建按版本号顺序执行的迁移（v1：schema_version/installs/generations/plugin_state），遇到比本程序更高的版本时拒绝打开；`AddInstall`（install_seq 递增）、`GetInstall`、`AddGeneration`、`CommitActive(plugin, expectedOld, new)`（同一事务内核对旧值、写入 committed_at 并切换 active_generation；旧值不一致返回 ErrStaleGeneration；代次必须属于该插件且未提交过）、`Active`、`SetEnabled`/`Enabled`、`LockPlugin`（按插件分别加互斥锁）。**与假设 6 的偏差：** 没有建 policy 占位表（目前没有使用方），由 T-03/T-04 通过迁移追加。
  - **过程中发现并修复的问题：** 第一版用可写 DSN 直接打开已有文件，测试 `TestCorruptStateNotRebuilt` 的「空文件」和「其他 SQLite 库」两个用例失败，报「状态库文件被改写」——驱动在连接时执行 `journal_mode(WAL)`，在 schema 检查之前就改写了文件头。改为先只读预检之后通过。Windows 下的盘符路径会被拼成 `file:///C:/...` 格式的 URI。
  - V-36（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 -v ./internal/oslock/ ./internal/core/state/` 全部 PASS：同一进程内第二个句柄得到 ErrLocked，解锁后锁文件仍存在且可以重新加锁；**跨进程**：子进程（重入测试二进制）持锁时父进程得到 ErrLocked，子进程退出（以 `cmd.Wait` 确认，没有用 sleep）后父进程可以加锁；**双 Core**：同一进程内第二个 Open 和另一进程中的 Core 都返回 ErrDataDirLocked，持有者退出后可以打开；PRAGMA 回读值为 `{wal 2 1 5000}`；关闭后重新打开，记录与 schema 版本都还在；**陈旧候选**：两个候选都以「无 Active」为前提时，后提交的那个得到 ErrStaleGeneration，Active 不变；已提交过的代次不能再次提交；不能提交其他插件的代次或不存在的代次；**损坏的状态库**：垃圾内容、空文件、其他 SQLite 库三种情况都拒绝打开，文件字节不变，且 core.lock 已释放；更高的 schema 版本被拒绝；32 个 goroutine 争用同一插件时并发度最大为 1，不同插件互不阻塞。全仓库 `go test -race ./...` 的 10 个包都 ok；`GOOS=windows/linux go vet ./internal/...` 通过；archtest 通过（新的允许边已生效）。
  - 剩余风险：windows 的 LockFileEx 与只读 URI 打开只经过交叉 vet，**没有运行**；Core 在首次初始化中途崩溃时会留下一个没有 schema 的文件，下次启动会拒绝并要求管理员处理（为了不把截断损坏的文件误当成新库，有意这样设计）；V-36 的「同插件并发 enable/update」「Core 被 kill 而子进程仍在写」留待 G1-15/G1-19 与 T-09。
  - 下一步：G1-13（帧编解码与 TCP/stdio 传输）。

### [x] G1-13 帧编解码与 TCP/stdio 传输
- 依据：PLAN T-02 第 3 步；CONTRACTS §1.1；WIRE §1；V-03。
- 内容：`internal/wire`：完整读取长度前缀和载荷，先校验长度再分配缓冲（1～4,194,304 字节），零长度、超限、截断、无法解码都返回明确错误；每个连接只有一个串行 writer；TCP 用 `net.Listen("tcp4","127.0.0.1:0")` 并持有 Listener；stdio 传输（stdout 只传帧，出现非协议字节就让会话失败）。
- 完成条件：同一份 fixture 在 TCP 和 stdio（用 `io.Pipe`/`os.Pipe` 模拟）上的往返结果一致；边界用例（0、4 MiB、4 MiB+1、截断、分多次 Read 才凑齐一帧）全部覆盖；并发写入的帧不交错（`-race`）。
- 回写：
  - 实际做了什么：`internal/wire/frame.go`：`MaxFrame=4<<20`；`ReadRaw` 先用 `io.ReadFull` 读满 4 字节前缀，**校验 1～MaxFrame 之后才分配缓冲**并读满载荷；流恰好在帧边界结束时返回 io.EOF，前缀或载荷中途结束返回 ErrTruncated，长度 0 返回 ErrEmptyFrame，超限返回 ErrFrameTooLarge；`ReadFrame` 解码失败返回 ErrDecode；`EncodeFrame` 在编码阶段拒绝空帧和超限帧（超限时不会写出任何字节）；`Writer` 先完整编码，再在锁内一次 Write，保证帧不交错。`internal/wire/transport.go`：`Conn`（唯一 reader：64 KiB 的 bufio，可在多个 goroutine 中安全调用的串行 writer，closer），`NewConn(r,w,closer)` 供 stdio 使用，`NewStreamConn` 供 TCP 使用，`ListenLoopback` 直接监听 `tcp4 127.0.0.1:0`，并核对确实是 loopback 且端口非零。stdout 混入日志时，日志字节会被当作长度前缀解析，读取失败（Envelope 的结构问题由 contracts/validate 继续拦截）。ARCHITECTURE 中 internal/wire 的状态已更新。
  - V-03（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 -v ./internal/wire/` 中 7 个测试（含 7 个子用例）全部 PASS：长度为 1 字节、恰好 4 MiB 的帧都能读出；0 → ErrEmptyFrame；**4 MiB+1 只提供前缀时就返回 ErrFrameTooLarge**（证明是在读取载荷之前拒绝的）；载荷截断和前缀截断都返回 ErrTruncated；空流返回 io.EOF；`iotest.OneByteReader` 逐字节读取时也能完整读出两帧；无法解码的载荷返回 ErrDecode；stdout 混入日志时读取失败；超限 Envelope 在编码时被拒绝，且写出 0 字节；**同一组 contracts golden 帧（4 个）分别经 TCP loopback 和 os.Pipe（stdio）往返，结果一致，重新编码后的字节也一致**；16 个 goroutine 各写 50 帧（总共 800 帧，载荷大小不同），读端每帧都能完整解码。变异检查：把 Writer 改为不加锁、分两次写前缀和载荷之后，测试报「读到损坏的帧（帧交错）」并失败；恢复后在 `-race -count=3` 下通过。archtest 通过（internal/wire 只 import contracts）。
  - 剩余风险：认证、Session 状态、优先通道、写超时还没有实现（G1-14/G1-16）；没有测试 Windows 上 stdio 管道的实际行为（只验证了交叉编译）。
  - 下一步：REVIEW-4。

### [x] REVIEW-4 集中检查（G1-10～G1-13）
- 检查清单同上；另外核对 T-01 停止条件、锁实现的平台 build tag，以及交叉编译是否仍然通过。
- 回写：
  - 检查过程与结论（commit d14cf01 之上）：
    1. 是否偏离 input.md：没有。新增依赖只是把已批准的 x/sys 从间接依赖改为直接依赖；新增模块 internal/oslock 已登记在 ARCHITECTURE 与 archtest 中。
    2. T-01 停止条件复核：G1-11 的结论仍然成立（契约能编译、生成稳定、示例成立、SQLite/PG 路径通过、版本已锁定）。
    3. 锁实现的平台 build tag：五平台 `go build` 与 `go vet` 全部 OK（windows 下编译的是 lock_windows.go，linux/darwin 下编译的是 `//go:build unix` 的 lock_unix.go）。
    4. `gofmt -l` 为空；`go mod tidy` 前后没有差异；`go test -race -count=3 ./internal/...` 的 4 个包都稳定 ok（没有偶发失败）。
    5. **发现缺陷（P1）**：`internal/core/state.preflight` 在拼 file URI 时给不以 / 开头的路径前面加了 /，用来处理 Windows 盘符路径，结果在 Unix 上把**相对路径**错误地变成了从根目录开始的绝对路径。实测：在临时目录下以相对路径 `reldata` 连续 Open 两次，第一次 `<nil>`，第二次报「状态库无法读取（可能已损坏，不会重建）: unable to open database file (14)」。影响：Core 用相对路径的 data_dir 首次启动之后就无法再启动（不会丢数据，但等于不可用）→ 登记 FIX-4。
    6. 设计提醒（不是缺陷，交给 G1-14）：`wire.Writer` 没有写超时，对端不读时 WriteFrame 会一直阻塞并持有锁；G1-14 的优先通道和「控制通道也无法推进时关闭 Session」必须在这一层之上实现写截止时间。
    7. 文档同步：docs/development.md 中仍写着「golang.org/x/sys 目前是间接依赖」，已过时 → 放进 FIX-4。
    8. 安全：oslock 锁文件权限为 0600，data_dir 为 0700；ReadRaw 在认证之前最多按协议上限分配 4 MiB（符合 CONTRACTS §1.1「先验证长度再分配受限缓冲」）。
  - 下一步：FIX-4（优先于 G1-14）。

### [x] G1-14 Session 调用引擎：配对、取消、准入与优先通道
- 依据：PLAN T-02 第 4 步；CONTRACTS §1.2；WIRE §4；V-05、V-42（T-02 分项）。
- 内容：`internal/wire` 中实现调用表，以 `(initiator_side, call_id)` 为键（每个 Session 一个实例）；call_id 单调递增；默认 deadline 30s，子请求 deadline 不晚于父请求；parent_call_id 只接受同 Session 内仍在途的 Core→Plugin 请求；准入预算为每个方向 256 个请求、每个方向 16 MiB 字节，另有一个跨 Session 共享的 64 MiB 全局预算对象，满额时立即返回 RESOURCE_EXHAUSTED；Response/Cancel/控制帧走有界优先通道；尚未写出的 Request 被取消时从队列撤出，已写出的 Request 才能随后发送 Cancel；迟到或重复的 Response 只计数并丢弃；payload 与 error 同时保留；控制通道也无法推进时关闭 Session，未完成的调用以 UNAVAILABLE 结束。
- 完成条件：两个方向使用相同 call_id 时互不串号；乱序 Response；业务槽满额时 Response/Cancel 仍能推进；发送前取消的请求从未出现在线上；所有测试在 `-race` 下通过。
- 回写：
  - 实际做了什么：新增 `internal/wire/session.go`（Core 与 SDK 共用，只靠 Options.Local 区分两端）。**调用表**：出站 `pending` 与入站 `inbound` 分表，各以 call_id 为键，等价于 `(initiator_side, call_id)`；出站 call_id 由 `nextID++` 单调递增，入站 call_id 必须严格递增，否则判为协议错误并关闭 Session。**期限**：默认 30s，取 Request.Deadline、ctx 与父调用三者中最早的；已经过期的直接返回 DEADLINE_EXCEEDED，不入队。**父调用**：插件侧的子请求只能引用仍在途的入站请求；Core 侧收到 parent_call_id 时，父调用必须是仍在途的出站请求，否则直接回 FAILED_PRECONDITION，不执行 Handler；子请求处理的 ctx 由父调用的 ctx 派生（父调用结束时子调用被取消），父调用的 `Attach` 以 `Inbound.ParentAttach` 交给子请求的 Handler，供 Core 派生可信上下文。**准入**：出站与入站各自限制 256 个、16 MiB，另有跨 Session 共享的 `Budget`（64 MiB）；任一项不足立即返回 RESOURCE_EXHAUSTED，入站超额时经优先通道回错误。**写出**：唯一的 writeLoop 每次先清空优先通道（Response/Cancel/控制帧/RESOURCE_EXHAUSTED），再取业务 Request；Request 的状态为 queued→writing→written 或 withdrawn。**取消**：本地期限到达或 ctx 取消时，先结束调用者，再按状态处理——queued 时撤回（永不上线、不发 Cancel），writing 时写完之后补发 Cancel，written 时立即发 Cancel。**响应**：迟到、重复或未知的 Response 只计入 LateResponses；未知 Cancel 计入 UnknownCancels。payload 与 ErrorInfo 一起交给调用者（返回 Response 的同时返回 *CallError）。**容错**：Handler panic 转为 INTERNAL；既没有结果也没有错误的响应补成 INTERNAL。**Session 失败**：优先通道写满（连控制帧都无法推进）、读写错误、Envelope 校验失败、session_id 不符时关闭 Session，未完成的出站调用以 UNAVAILABLE 结束，入站处理被取消。控制帧与 EVENT 在读循环中同步交给 Options.Control（供 G1-17 保证 ACTIVATED 先于后续业务帧生效）。
  - V-05 / V-42（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=10 ./internal/wire/` 全部 ok，覆盖了 17 个 Session 测试：**两个方向同时使用 call_id=1 时不串号**；后发先回（3→2→1）的结果各归各；payload 与 error 同时保留；panic 转为 INTERNAL；**256 个在途时第 257 个在 100ms 内被本地拒绝**（对端入站上限放大到 1000，排除了对端拒绝的可能），满额时 Cancel 仍能送达对端，Response 仍能返回，释放后可以再次准入；对端入站上限为 2 时第 3 个收到 RESOURCE_EXHAUSTED；每方向字节预算与跨 Session 全局预算都生效，完成后全局占用回到 0；**发送前取消的请求 2 从未上线**（对端只看到 1 和 3，UnknownCancels=0）；本地超时后对端收到 Cancel，不合作的 Handler 之后返回的 Response 计入 LateResponses=1，没有再次完成调用；A→B→A 子调用能继承父调用的 Attach，期限不晚于父调用；父调用超时后子调用被取消；父调用已在 Core 侧结束（取消）时，子请求被 Core 拒绝且不执行 Handler；插件侧引用不存在的父调用时本地拒绝；期限已过的调用不上线；断链后在途调用为 UNAVAILABLE；重复的 call_id、其他 Session 的帧都会关闭 Session；**写出卡住且优先通道写满时关闭 Session，在途调用以 UNAVAILABLE 结束，没有死锁**；生产常量与 CONTRACTS §1.2 一致。变异检查：(1) 去掉写循环中跳过 withdrawn 的逻辑后，对端收到 [1 2 3] → 测试失败；(2) 去掉出站准入后，`TestAdmissionFull…` 与 `TestByteBudgets` 失败；恢复后通过。全仓库 `go test -race ./...` 没有失败的包。
  - 过程中修正的问题：(a) 第一版 `TestCoreRejectsChildOfFinishedParent` 让父调用通过超时结束，插件侧子请求被钳到已过期的父期限、在本地就超时了（这是正确行为），改为让 Core 主动取消父调用来构造场景；同时新增「期限已过不入队」。(b) 第一版满额测试 `-count` 多跑时会偶发失败：被取消的那一条可能还在队列里就被撤回，测不到 Cancel 的推进；改为先确认它已经写出，修改后定向跑 30 次、整包跑 10 次都稳定通过。
  - 剩余风险：没有写超时的底层连接，只能依靠「优先通道写满时关闭」这一条来脱困：对端停止读取但通道还没写满时，业务调用会一直等到各自的期限；单个对端停止读取时会让 Session 内的所有调用超时，这符合规格中「关闭 Session」的处理边界。准入以 payload 字节计量，不含 Envelope 其余字段。业务 Request 与阶段门禁（Preparing 阶段拒绝业务）在 G1-17 接入。
  - 下一步：G1-15（Native 进程启动、runtime.lock 与进程树）。

### [x] G1-15 Native 进程启动、runtime.lock 与进程树
- 依据：PLAN T-02 第 2 步；CONTRACTS §7.1、§12.2；V-06、V-36（分项）。
- 内容：`internal/core/process`：根据安装记录取得入口程序，生成 256-bit Session Secret，构造精简环境（不传 Core 管理凭证），Unix 下使用 Setpgid 并对进程组发信号，Windows 下用 Job Object（build tag）；收集 stdout/stderr 日志并标记 plugin_id；取得退出状态。`sdk` 在任何用户回调之前获取 `<plugins>/<id>/runtime.lock`，锁文件不删除；Core 可以在旧进程退出后临时持有该锁，再释放。
- 完成条件：测试用重入的辅助进程验证：日志带插件 ID、环境中没有管理凭证、杀进程组后子进程也被回收（darwin 实测）、runtime.lock 被占用时第二个 SDK 实例启动失败、Windows 代码交叉编译通过（运行状态标为「未运行」）。
- 回写：
  - 实际做了什么：`internal/wire/env.go` 定义 Core 与 SDK 之间的启动环境约定（ALONABOT_PLUGIN_ID/SESSION_SECRET/TRANSPORT/CORE_ADDR/BUILD_DIGEST/RUNTIME_LOCK、tcp/stdio、runtime.lock 文件名），并写入 WIRE §5A。新增 `internal/core/process`：`NewSecret`（crypto/rand 生成 32 字节，十六进制）；`Entrypoint`（经 manifest.ValidateRelPath 校验后拼接安装目录，要求是普通文件，Unix 下还要求可执行）；`Start`：环境变量**只从白名单**继承（PATH/HOME/TMPDIR/TEMP/TMP/SystemRoot/LANG/TZ），再叠加会话变量；stdout/stderr 由**自建的 os.Pipe** 接管，写端以 *os.File 交给子进程，父进程在 Start 之后关闭自己手里的写端，因此 cmd.Wait 不会被仍持有管道的孙进程卡住；日志按行转发并标记 plugin_id 与来源流，超过 64 KiB 的行分块完整转发；stdio 模式下 stdin/stdout 作为协议管道交给调用方，只有 stderr 进入日志；`Exited`/`ExitState`/`WaitLogs`；进程树：Unix 使用 `Setpgid` 后对整个进程组发 SIGKILL（组内已无进程时 ESRCH 视为成功），Windows 使用 Job Object（KILL_ON_JOB_CLOSE；Start 之后再 AssignProcessToJobObject，kill 时 TerminateJobObject，Release 时关闭句柄）；`LockRuntime(pluginDir)` 供 Core 临时持有 runtime.lock。`sdk/runtime.go`：`acquireRuntimeLock` 从 ALONABOT_RUNTIME_LOCK 读取路径（必须提供且必须是绝对路径），取不到锁时拒绝启动，供 G1-16 的 Run 在任何用户回调之前调用。
  - V-06 / V-36（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=3 ./internal/core/process/` 与 `go test ./sdk/` 都 ok：在父进程中设置 ALONABOT_ADMIN_TOKEN 和 OTHER_PLUGIN_SECRET 之后，子进程环境中都没有这两个变量，且只有白名单变量与会话变量；日志行形如 `demo|stdout|…`、`demo|stderr|…`，100 KiB 的长行被分块完整收到；退出码 3 能正确取得；stdio 管道往返正常，stdout 内容不进入日志、stderr 进入日志；**Unix 下杀掉进程组后，孙进程（另一个 pid）也被回收**；对已经为空的进程组再次 KillTree 没有副作用；Entrypoint 拒绝 `../`、绝对路径、目录、不存在的文件、不可执行的文件；Core 持有 runtime.lock 时其他持有者失败，释放后锁文件仍然保留；SDK 侧第二个实例因 ErrLocked 拒绝启动，前一个释放后可以取得锁，缺少环境变量或路径为相对路径时都失败。`GOOS=windows/linux go vet ./...` 通过（Windows 的 Job Object 路径只验证了编译，**未运行**）；全仓库 `go test -race ./...` 的 12 个包都 ok。
  - **过程中发现并修复的问题：** 第一次变异检查（把 KillTree 改成只杀领头进程）时测试没有失败——辅助进程用 `select{}` 挂起，而进程里没有其他 goroutine，Go 运行时判定死锁后立即退出，孙进程本来就会自己消失，**测试是空转的**。改为 `time.Sleep(time.Hour)` 之后，同样的变异让测试失败（「孙进程 55283 在进程组被杀后仍存活」），恢复后通过；变异运行遗留的孙进程已用 pkill 清理，`pgrep` 检查无残留。另外 Unix 专用的进程组测试拆到 `process_unix_test.go`（`//go:build unix`），修复了 Windows 交叉 vet 时报出的 `syscall.Kill` 未定义。
  - 剩余风险：Windows 下进程是在 Start 之后才加入 Job 的，加入之前创建的孙进程不受管理（代码中已用 `shortcut:` 注释标明，需要时改为挂起创建再加入）；合作型进程树无法阻止恶意 detach（setsid）；runtime.lock 只在 sdk 内实现了获取，Run 调用它的路径在 G1-16 才接上。
  - 下一步：G1-16（握手、Bootstrap、REGISTER 与 PREPARED）。

### [x] G1-16 握手、Bootstrap、REGISTER 与 PREPARED
- 依据：PLAN T-02 第 5 步；CONTRACTS §1.1、§4.2；WIRE §3；V-04、V-37（分项）。
- 内容：Core 侧校验 Hello（Secret 恒时比较、plugin_id、build_digest、协议协商）后发出 HelloAck（含 Bootstrap：generation_id、原始配置 JSON（只检查是否为 JSON 对象及大小）、DB 描述、阶段预算）；SDK 侧按顺序执行配置准备（B1 中只有原始 JSON 透传和可选回调的占位）→ OnStart → 发送 REGISTER；Core 原子校验整份注册集（重复 route ID、超出 4 MiB 时返回 RESOURCE_EXHAUSTED），失败时不留下注册残留；SESSION_ERROR 携带失败阶段；握手 10s、OnStart 60s 的阶段 watchdog（测试中注入缩短的预算）；Bootstrap 内容不写入日志。
- 完成条件：测试覆盖：伪造 Secret、跨插件 Secret、协议 Major 不一致、未知必需能力、REGISTER 超限、重复 route、OnStart 返回错误、OnStart 挂起超过预算；日志中找不到 Secret/Bootstrap 内容。
- 回写：
  - 实际做了什么：**contracts/validate/register.go**：`Register` 对整份注册做原子校验（route_id 非空且不重复，重复时返回 ALREADY_EXISTS；kind 合法并与 spec 分支一致；candidate_config_json 必须是 JSON 对象；HttpListener 只接受 127.0.0.1 或 ::1 的非零端口），另外提供 `IsJSONObject`。**internal/core/host/prepare.go**：`Prepare(ctx, conn, Expect, Bootstrap, Limits)`，用阶段 watchdog 在超时时关闭连接，握手阶段用 Handshake 预算，准备阶段用 Config+Migration+OnStart 之和作为硬截止，并写入 Bootstrap 的 preparation_deadline。首帧必须是 HELLO，认证在 `authentic` 中完成（Secret 用 `subtle.ConstantTimeCompare`，同时核对插件 ID 与 build_digest；任一不符都**直接关闭、不回复，也不区分是哪一项**）；认证通过后才分配 16 字节随机 session_id，再调用 `validate.Negotiate`，失败时发 SESSION_ERROR(HANDSHAKE, FAILED_PRECONDITION)；Bootstrap 只检查配置是否为 JSON 对象，HELLO_ACK 超过 4 MiB 时发 SESSION_ERROR(RESOURCE_EXHAUSTED)，拒绝启动；准备阶段只接受 REGISTER 或插件报告的 SESSION_ERROR（后者原样返回插件报告的阶段与错误码），其他帧（例如业务 REQUEST）返回 PROTOCOL_ERROR；收到超过 4 MiB 的帧按 RESOURCE_EXHAUSTED 处理；REGISTER 校验失败时发 SESSION_ERROR(REGISTER) 且不返回任何结果；成功后回 PREPARED（route hash 与 config digest）。`DefaultLimits` 与 CONTRACTS §4.2 一致。整个函数不写日志，返回的错误中不包含 Secret 或配置内容。**internal/wire**：新增 `Conn.WriteFrameTimeout`（底层连接支持 SetWriteDeadline 时生效）。**sdk/session.go**：`readLaunch`（校验环境变量）、`dial`（TCP 地址必须是 loopback；stdio 使用 os.Stdin/os.Stdout）、`clientHandshake`（发送 HELLO，10s 内等待 HELLO_ACK）、`prepare`（配置原样透传（B1 占位）→ 在独立 watchdog 下执行 OnStart，返回错误、panic、超时都会发 SESSION_ERROR(ON_START)，其中安全消息不含用户错误原文，原文只写 stderr → 发送 REGISTER，超限时发 SESSION_ERROR(REGISTER, RESOURCE_EXHAUSTED) → 等待 PREPARED，或处理 Core 的拒绝），`buildRegister` 中的命令路由使用临时 route_id `cmd.<路径>`（代码中有 `shortcut:` 注释，T-04 替换）。`sdk.Run` 依次执行 Collect → 声明了 HTTP 或调度时返回 UNSUPPORTED（T-07）→ acquireRuntimeLock → readLaunch/dial → 握手 → prepare，之后仍返回 ErrNotImplemented（激活在 G1-17）。WIRE §5A 补充了握手/准备阶段的失败语义与 PREPARED 摘要的定义。
  - V-04 / V-37（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go test -race -count=1 -v ./internal/core/host/`（16 个用例）与 `./sdk/`（新增 9 个）全部 PASS：**伪造 Secret、使用其他插件的 Secret、插件 ID 不符、build digest 不符**四种情况都返回 UNAUTHENTICATED，且插件侧读不到任何帧（Core 没有回复），错误信息中不含 Secret；Major 不一致、未知必需能力都会收到 SESSION_ERROR(HANDSHAKE, FAILED_PRECONDITION)；**重复 route** 收到 SESSION_ERROR(REGISTER, ALREADY_EXISTS)，Prepare 返回 nil 结果（没有注册残留）；**超过 4 MiB 的 REGISTER**：Core 侧读到长度前缀就判为 RESOURCE_EXHAUSTED，SDK 侧在编码阶段就报 SESSION_ERROR(REGISTER, RESOURCE_EXHAUSTED)；Bootstrap 超限时返回 RESOURCE_EXHAUSTED；**OnStart 返回错误**时 Core 得到 ON_START/INTERNAL，SESSION_ERROR 的安全消息中不含用户错误里的 `SECRET-XYZ`；**OnStart 挂起超过预算**（忽略 ctx，预算 100ms）时，SDK 在预算到达后报告 DEADLINE_EXCEEDED，Core 的准备阶段超时、握手超时都约在预算内返回；准备期收到业务 REQUEST 时返回 PROTOCOL_ERROR；所有失败的错误信息中都不包含 Secret 或配置值 `CFG-SECRET-VALUE`；HELLO 字段（空 session_id、source_plugin_id、secret、digest、major）正确；SDK 拒绝非 loopback 的 Core 地址、过短的 Secret、未知的传输方式；Run 遇到定时任务声明时返回 UNSUPPORTED。三平台 `go vet` 通过；全仓库 `go test -race ./...` 的 13 个包都 ok；archtest 通过（host 只 import contracts 与 internal/wire）。
  - **过程中发现并修复的问题：** 第一版测试卡死——「准备期业务帧」和「防泄漏」两个用例中假插件不读数据，Core 在 `sendErr` 中写 SESSION_ERROR 时，同步的 net.Pipe 让写操作永久阻塞（准备阶段预算 15 分钟，watchdog 也来不及解救）。这反映了真实的缺陷：**对端不读时，Core 的失败路径会被卡住**。修复办法是新增 `WriteFrameTimeout`，失败通知限时 1s 后关闭连接；修复后这两个用例都在约 1s 内结束。
  - 剩余风险：SDK 中配置准备只是透传（OnLoad、默认值、Validate 在 T-08 实现）；route_id 是临时规则；Windows 管道是否支持 SetWriteDeadline 不确定，不支持时会退化为普通写，失败路径依赖 Core 随后终止进程树来解除阻塞；`sdk.Run` 中「OnStart 超时后退出进程」与激活流程留给 G1-17/G1-18；**真正的跨进程端到端**（Core host 与 SDK 在不同进程中）留给 G1-19。
  - 下一步：REVIEW-5。

### [x] REVIEW-5 集中检查（G1-14～G1-16）
- 检查清单同上；重点检查并发与竞态（`-race` 跑 10 次：`go test -race -count=10` 覆盖 wire 和 session 相关包）、错误路径上的资源泄漏（goroutine、fd、Listener），并用 code-review-expert 审查。
- 回写：
  - 流程说明：code-review-expert 的 SKILL.md 第 7 步要求先确认再修改；按 goal-mode 的 SKILL.md「红线」（禁止提问），审查发现直接登记为 FIX task。
  - 检查过程与结论（审查范围 `c769f5a..e81ace6`，即 G1-14～G1-16 与 FIX-4）：
    1. 是否偏离 input.md：没有；没有新增依赖。
    2. **并发**：`go test -race -count=10 ./internal/wire/ ./internal/core/host/ ./internal/core/process/ ./sdk/` 4 个包全部 ok，没有数据竞争，也没有偶发失败。
    3. **资源泄漏（临时探针，已删除）**：Session 建立/调用/关闭 200 对，加上 TCP 上的 50 个 Session：goroutine 2→2，fd 5→5；host.Prepare 认证失败 200 次：goroutine 2→2（watchdog 都已退出）；process 启动并回收 50 次（TCP 模式）：goroutine 2→2，fd 5→5；stdio 模式 20 次：fd 5→5，但这依赖于 `os.File` 的 finalizer 在 GC 时关闭管道端，见 P3-2。
    4. code-review-expert 审查（COMMENT，没有 P0/P1）：
       - **P3-1 控制帧 Envelope 的构造重复了三份**：`wire.Session.SendControl`、`host.controlEnvelope`、`sdk.controlEnv` 内容几乎一致，日后字段改动时容易出现分歧。修复：在 `internal/wire` 中提供 `ControlEnvelope(sid, minor, sourcePluginID, kind, msg)`，三处复用（archtest 允许 host 和 sdk 都 import wire）。
       - **P3-2 `process.Release` 没有关闭 stdio 模式下 Core 持有的 Stdin/Stdout**：目前靠 GC 时的 finalizer 关闭，而 Core 长期持有 Process 时会推迟关闭。修复：Release 时显式关闭它们（重复关闭不报错）。
       - **P3-3 SDK 在 Bootstrap 缺少 PhaseLimits 时把 OnStart 预算视为 0**，会报出一个误导性的「超时」。修复：PhaseLimits 缺失、OnStart 预算为 0 或 preparation_deadline 缺失时，以 PROTOCOL_ERROR 报告 SESSION_ERROR(CONFIG)。
       - 记录但不修：`host.Prepare` 函数较长（约 150 行），按步骤顺序书写、可读性尚可，G1-17 加入激活后再评估是否拆分；SDK 在 REGISTER 中固定写 `MigrationState{NONE}`，属于 B1 已知限制，T-08 接入迁移后修正（不在 SDK 主包 import gormdb，以免所有插件都被迫依赖 gorm）。
    5. 安全：Secret 恒时比较、认证失败时不回复细节、失败路径的安全消息不含用户错误原文与配置内容，都已有测试覆盖；没有发现注入或越权的路径。
  - 下一步：FIX-5（优先于 G1-17）。

### [x] G1-17 提交与激活门禁
- 依据：PLAN T-02 第 6 步；CONTRACTS §4.1、§12.1；V-21、V-37（分项）。
- 内容：PREPARED 之后，Core 在状态库事务中提交 ActiveGeneration+COMMITTED（事务内不等待 RPC），然后发送 ACTIVATE；SDK 执行 OnReady（30s 预算），由唯一 writer 写出 ACTIVATED 后打开 ready gate；Core reader 在应用 Activated 状态之后才处理后续 REQUEST；Preparing/Prepared/Activating 阶段收到业务 REQUEST 一律拒绝；提交后 OnReady 失败进入 Degraded，不回到旧代次；普通 restart 复用原有提交记录和 Generation。
- 完成条件：测试覆盖：激活前的业务 REQUEST 被拒绝；ACTIVATED 与紧随其后的 REQUEST 在同一连接上按顺序处理；OnReady 失败后状态为 Degraded 且状态库中仍是新代次；restart 不产生新 Generation。
- 回写：
  - 实际做了什么：**internal/wire**：`Options.Admit`——读循环在执行 Handler 之前**同步**调用的业务门禁，不持锁；`SendControlSync`/`SendEnvelopeSync`——优先通道中的帧带写出通知，等唯一 writer 写完后才返回；把 `NewSession` 拆成 `New` + `Start`（原 `NewSession` 仍等于两者组合）。**internal/core/host/activate.go**：`Activate(ctx, conn, prep, ActivateOptions)`。新候选先调用 `Commit`（由调用方传入 `state.CommitActive`，事务内只有元数据，不等待 RPC）；提交失败时发送 SESSION_ERROR(ACTIVATE)、关闭连接、**不发送 ACTIVATE**，旧代次保持权威；`Commit=nil` 表示普通 restart，复用原有提交记录与 Generation。之后创建 Session，`admit` 在 Activated 之前一律回 FAILED_PRECONDITION；`control` 在读循环中同步处理 ACTIVATED（核对代次与 ready 之后切换为 Activated）与 SESSION_ERROR；发送 ACTIVATE（generation_id、commit_marker 为代次+route hash+config digest 的 SHA256、OnReady 截止时间）；等待 ACTIVATED、OnReady 超时、断链或 ctx 取消。失败时用 CAS 从 Activating 转为 **Degraded**（ACTIVATED 与超时同时到达时以 Activated 为准），状态库中仍是新代次。**sdk/activate.go**：PREPARED 之后用 `wire.New` 创建 Session，**赋值完成后再 Start**；`admit` 在 ready gate 打开之前拒绝业务；收到 ACTIVATE 时核对代次与 Bootstrap 一致，然后在独立 goroutine 中执行 OnReady（期限取 ACTIVATE 给出的 deadline，由 watchdog 兜底，panic 被 recover），失败时发送 SESSION_ERROR(ON_READY, PLUGIN_NOT_READY/DEADLINE_EXCEEDED)（安全消息不含用户错误原文）；成功后用 `SendControlSync` 写出 ACTIVATED，**写完才**设置 ready 并关闭 readyCh；ready 之后的业务 Handler 暂时一律回 UNSUPPORTED（代码中有 `shortcut:` 注释，T-03/T-04 接入分发）。`sdk.Run` 在 prepare 之后调用 activate，然后服务到 Session 结束（STOPPING 留给 G1-18）。
  - V-21 / V-37（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：host 新增 6 个测试、sdk 新增 4 个测试，都 PASS：**ACTIVATE 之前状态库已提交 gen-1**；**Activating 阶段的业务 REQUEST 被拒绝**（FAILED_PRECONDITION）；**ACTIVATED 与紧随其后的 REQUEST 在同一次 Write 中送达时，REQUEST 被接受**（Core 先应用 Activated）；**OnReady 失败（SESSION_ERROR(ON_READY)）与 OnReady 超时（150ms）都进入 Degraded，状态库中仍为 gen-1**；**restart（Commit=nil）使用原代次 gen-1，不再提交**（随后再次提交 gen-1 失败，证明 restart 路径没有提交）；陈旧候选提交失败时收到 SESSION_ERROR 而不是 ACTIVATE，返回 ErrStaleGeneration，Active 保持 gen-0；SDK 侧 ready 之前的请求被拒绝；**同步 net.Pipe 下 Core 读走 ACTIVATED 之前 activate 不返回（ready gate 未打开）**，读走之后才返回，之后的请求能通过门禁；OnReady 错误与挂起分别报告 PLUGIN_NOT_READY 与 DEADLINE_EXCEEDED；代次不符的 ACTIVATE 使激活失败。变异检查：(1) Core 侧把 Activated 的切换改为异步之后，「紧随 ACTIVATED 的 REQUEST」被拒绝，测试失败；(2) SDK 侧把 ready gate 移到 ACTIVATED 写出之前打开，测试报「ACTIVATED 尚未写出时 activate 不应返回」；恢复后都通过。`go test -race -count=5 ./sdk/ ./internal/core/host/ ./internal/wire/` ok；全仓库 13 个包都 ok；三平台 vet 通过。
  - **过程中发现并修复的问题：** `-race` 报出数据竞争——SDK 在 `NewSession` 返回之后才给 `a.sess` 赋值，而读循环已经启动，ACTIVATE 先到时 `runOnReady` 会读到尚未赋值（最坏为 nil）的 Session。修复办法是把 Session 拆成 New/Start，先赋值再启动；Core 侧的 admit/control 不访问 Session 指针，不受影响。
  - 剩余风险：Degraded 只记录在内存中（Instance 状态），没有持久化，Core 重启后需要依据状态库与进程情况重新判断（T-09 恢复编排）；STOPPING/OnStop/断链退出与崩溃计数在 G1-18；SDK 业务分发未实现。
  - 下一步：G1-18（停止、阶段超时与有限恢复）。

### [x] G1-18 停止、阶段超时与有限恢复
- 依据：PLAN T-02 第 7 步；CONTRACTS §4.2；V-06、V-21（分项）。
- 内容：停用时冻结新的根调用，最多排空 6s（父链内合法的子调用仍可完成）→ Cancel 剩余调用 → 发送 STOPPING，deadline 为 `min(t0+10s, 发送时刻+4s)` → SDK 用独立 context 执行 OnStop → 到 t0+10s 时终止进程树；Core Shutdown 并行停止所有插件，共享同一个截止时间；EOF 时旧 Session 立即失效，SDK 收到 EOF 后退出；5 分钟内第三次异常退出时自动禁用，正常退出不重启，手动启用会重置失败窗口；用户函数不响应 context 时由 watchdog 结束进程。
- 完成条件：测试（注入缩短的时间常量，另有测试核对生产常量等于契约数值）覆盖：drain 期间子调用完成、STOPPING deadline 公式、OnStop 挂起时进程在总截止时间被杀、多个插件并行停止的总耗时不超过一份预算、崩溃计数与自动禁用、正常退出不重启。
- 回写：
  - 实际做了什么：**internal/wire**：抽出 `withdrawOrCancel`（abandon 复用）；新增 `CancelOutbound(reason)`（外部结束全部本端在途请求：调用者得到 CANCELLED，对端按写出状态收到 Cancel 或请求被撤回）、`Inflight()`、`WaitIdle(ctx)`（在途数变化时关闭并替换 changed 通道来通知，不轮询）。**修复**：`Admit` 原来在父调用查找之前执行，看到的 `Inbound.ParentCallID` 恒为 0，Draining 时无法放行合法子调用；改为先填入帧中声明的 parent_call_id，有效性仍由随后的查找校验。**internal/core/host**：新增 Draining/Stopping/Stopped 状态，Draining 时 admit 只放行带 parent 的子调用，新根调用回 UNAVAILABLE；`Instance.Stop(t0, reason, tree, StopTiming)`：冻结 → `WaitIdle` 最多 Drain → `CancelOutbound` → 发送唯一 STOPPING，截止为 `min(t0+Total, 发送时刻+OnStopAt)` → 等待 `tree.Exited()`，到 t0+Total 时 `KillTree()`，并且**以 Exited 确认退出**（不用 sleep）→ 关闭 Session → Stopped；`StopAll` 让所有实例用同一个 t0 并行停止；`RestartPolicy`（窗口 5 分钟、限 3 次、正常退出返回 DecisionNone、窗口内第 3 次异常退出返回 DecisionDisable、`Reset` 用于手动启用，窗口滑动，不同插件之间隔离；失败窗口只在内存中，禁用由调用方经 state.SetEnabled 持久化）；`DefaultStopTiming` = 6s/10s/4s。**sdk**：activation 增加 lifetime 上下文；收到 STOPPING 时在读循环中同步关闭 ready gate，之后 `shutdown(STOPPING 截止)`：取消 lifetime → 在**全新的** context（期限即 STOPPING 截止）中执行 OnStop，watchdog 到期即放弃等待（panic/错误只写 stderr）→ 关闭 Session；重复的 STOPPING 视为协议错误；`serve`：收到 STOPPING 后正常返回；**断链/EOF** 时在 `onStopBudget`=4s 内执行 OnStop，然后返回错误（不重连）；`Run` 改为 `activate` + `serve`，返回即意味着进程退出（用户函数挂起不会阻止退出）。
  - V-06 / V-21（T-02 分项）证据（darwin/arm64，go1.27.1，本 task 的 commit）：host 新增 6 个测试、sdk 新增 5 个测试，全部 PASS（时间常量注入为 Drain 200ms/Total 600ms/OnStop 300ms）：**排空期间父链内的子调用（parent=在途请求）完成，新根调用被拒绝（UNAVAILABLE）**，在途调用正常完成后提前结束排空（Drained=true），无需强杀；**STOPPING 截止 = min(t0+Total, 发送时刻+OnStop)**，两个分支都有覆盖（排空完成时取 发送时刻+300ms，排空超时、Total=300ms 时取 t0+Total，误差 ≤1ms）；排空超时后对端收到在途请求的 CANCEL，调用者得到 CANCELLED；**插件不退出时在 t0+Total（600ms，实测略多于 600ms）KillTree**；**3 个插件并行停止的总耗时不超过一份 Total（另加 300ms 余量）**；重启策略：正常退出不重启，前两次异常退出重启，5 分钟内第三次异常退出时禁用，超出窗口的失败不计入，Reset 后重新计数，插件之间互不影响；生产常量等于 6s/10s/4s 与 5 分钟内 3 次；SDK：OnStop 的 ctx 一开始未取消，截止等于 STOPPING 截止，lifetime 已取消，STOPPING 之后的请求被拒绝（FAILED_PRECONDITION），之后 serve 正常返回；OnStop 挂起时在截止（150ms）附近返回；**断链后执行 OnStop，serve 返回错误，lifetime 已取消**；重复 STOPPING 会关闭 Session；SDK 预算为 OnStop 4s、握手 10s。变异检查：(1) 把 StopAll 改为串行后耗时 1.80s，超过 600ms 预算，测试失败；(2) STOPPING 截止不再受 t0+Total 约束时，截止落在 t0+Total 之后，测试失败；恢复后通过。`go test -race -count=3 ./internal/core/host/ ./sdk/ ./internal/wire/` ok，全仓库通过，三平台 vet 通过。
  - **过程中的失误（已纠正）：** 写 stop_test 时第一次误用了**主仓库路径**（`alonabot/internal/core/host/stop_test.go.tmp`，不在 worktree 中），发现后立即删除该文件及其产生的空目录，主仓库 `git status` 恢复原状（仍只有一个原本就存在的未跟踪 `.claude/`，我没有动它），之后改在 worktree 路径下写入。
  - 剩余风险：失败窗口与 Degraded 只在内存中，Core 重启后不保留（T-09 恢复编排需要考虑）；异常/正常退出的判定（abnormal）由调用方依据「是否已发 STOPPING、退出码」给出，G1-19 在装配中实现；真实进程上的 STOPPING→退出、强杀与 runtime.lock 释放在 G1-19 验证。
  - 下一步：G1-19（cmd/alonabot 装配与真实进程验收切片）。

### [x] G1-19 cmd/alonabot 装配与真实进程验收切片
- 依据：PLAN T-02 第 8 步与「对照现有验收」。
- 内容：`cmd/alonabot` 提供开发用的启停入口；`examples/plugins/minimal` 以真实子进程分别在 TCP 和 stdio 上跑完整生命周期；故障辅助进程覆盖挂起、错帧、stdout 日志污染、迟到响应、突然退出；场景包括双 Core、同插件并发操作、Core 被 kill 后 SDK 断链退出；检查进程和 Listener 都已关闭。把实际命令写入 docs/development.md。
- 完成条件：所有场景测试通过并可以重复运行（`-count=3`）；整理 V-03/V-04/V-05/V-06/V-21/V-36/V-37/V-42 的 T-02 分项证据表，明确哪些留待 T-07/T-08/T-09/T-10 接入；Linux/Windows 的运行状态记为「未运行」。
- 回写：
  - 实际做了什么：**internal/core/supervisor**：`Launch(ctx, store, Spec)` 在逐插件锁内依次完成：生成 Secret → 设置环境变量（插件 ID、Secret、build digest、runtime.lock 绝对路径、传输方式）→ TCP 时 `ListenLoopback`（Core 持有到停止）或使用 stdio 管道 → `process.Start`（工作目录为插件目录）→ 握手期限内 Accept（插件进程提前退出时立即放弃）→ `host.Prepare` → `host.Activate`（新候选的 Commit 为 `store.CommitActive`）；任何一步失败都会 KillTree，并**等待 Exited 确认退出**后再释放资源。`Running.Stop`（`Instance.Stop` → 关闭 Listener → 释放进程）、`StopAll`、`WaitExit`（返回是否为异常退出；异常退出时立即关闭旧 Session）。**cmd/alonabot**：`alonabot run --data-dir --plugin-id --entry [--stdio]`：`state.Open`（同时取得 core.lock）→ 开发模式下登记安装记录（来源 dev-local，摘要为可执行文件的 SHA256）→ 首次运行新建并提交代次（mode=new），之后复用 Active 代次（mode=restart）→ SIGINT/SIGTERM 时执行 Stop，异常退出时交给 `RestartPolicy` 决定重启或禁用（禁用会经 SetEnabled 持久化）。**examples/plugins/lifecycle**：可以真实运行的 SDK 插件（OnStart/OnReady/OnStop 以及一个命令，日志只写 stderr）；`minimal` 保留为签名组合示例（它声明了 HTTP/调度，按设计返回 UNSUPPORTED）。**故障辅助进程**（supervisor 测试中重入测试二进制，直接收发原始协议帧）：none/nostop/crash/hang/badframe/pollute/exit-early。docs/development.md 补上了 run 用法与验收命令，ARCHITECTURE 中 cmd/supervisor/examples 的状态已同步。
  - **过程中发现的问题：** (a) 第一次冒烟用 shell 后台进程加轮询，脚本写错导致卡住；残留的 alonabot/lifecycle/zsh 进程已全部清理，改用 Go 端到端测试。(b) `-race` 下「合作插件被判为强杀」：调试输出显示插件在收到 STOPPING 的同一毫秒就调用了 `os.Exit(0)`，但 -race 编译的辅助进程默认会在退出前 sleep 1s（GORACE `atexit_sleep_ms` 默认为 1000），超过了测试注入的 800ms 总截止。这不是产品缺陷，测试中为辅助进程设置 `GORACE=atexit_sleep_ms=0`。(c) Windows 交叉 vet 时报 `syscall.Kill` 未定义：两个依赖 Unix 信号与进程组的测试文件加上 `//go:build unix`。
  - **T-02 真实进程验收**（darwin/arm64，go1.27.1，本 task 的 commit；命令 `go test -count=3 ./cmd/alonabot/ ./internal/core/supervisor/` 与 `go test -race -count=3 ./internal/core/supervisor/` 都 ok，结束后 `pgrep` 确认没有残留进程）：
    - 端到端（真实 alonabot 进程 + 真实 SDK 插件进程）：**TCP 与 stdio** 都走完 OnStart → OnReady → 已激活(mode=new) → SIGTERM → OnStop → 「lifecycle: 正常退出」，「强杀=false」，alonabot 退出码 0，插件 pid 消失；**restart** 复用原代次（mode=restart，generation 相同）；**同一 data_dir 的第二个 Core** 以非零退出码退出，并报告「已被另一个 Core 锁定」；**Core 被 SIGKILL** 之后，插件 pid 在断链预算内退出，随后新的 Core 能取得锁，并以 restart 复用原代次恢复运行（证明 core.lock 与 runtime.lock 都已释放）。
    - supervisor（故障辅助进程）：TCP/stdio 正常停止时退出码为 0、没有强杀、**停止后 Listener 不再接受连接**、状态库提交了 gen-1；不响应 STOPPING 的插件在总截止（800ms）被终止；激活后崩溃（退出码 3）被报告为异常退出，有限重启的决定依次为 Restart、Restart、Disable；**连上后挂起**→DEADLINE_EXCEEDED、**错帧**→PROTOCOL_ERROR、**stdout 日志污染（stdio）**→失败、**连接前退出**→「连接前退出」，四种情况都被终止，且不提交候选（Active 为空）；**同一插件并发启动（真实 SDK）**：第二个实例因 runtime.lock 拒绝启动（日志中可以看到），第一个实例不受影响。
  - **T-02 证据表（V 分项；Linux/Windows 全部「未运行」，darwin/amd64 只经 Rosetta 跑过单元测试，没有跑端到端）：**
    | V | T-02 已取得的证据 | 留待 |
    | --- | --- | --- |
    | V-03 | 帧 1～4 MiB 边界、零长度/超限/截断/错帧（G1-13）；同一 fixture 在 TCP 与 stdio 上往返一致；真实进程在两种传输上的完整生命周期（G1-19） | T-10 五平台 |
    | V-04 | 伪造 Secret、跨插件 Secret、插件 ID 或 digest 不符时不回复并关闭（G1-16）；其他 Session 的帧、代次不符的 ACTIVATED 被拒绝（G1-14/G1-17） | — |
    | V-05 | 两个方向使用相同 call_id、乱序、迟到/重复 Response 只计数、父调用结束后子调用被拒绝或取消（G1-14） | T-07 Scheduler 的迟到响应释放运行槽 |
    | V-06 | 日志带插件 ID（G1-15）；stdio stdout 污染使会话失败（G1-13/G1-19）；5 分钟内第 3 次异常退出时禁用，正常退出不重启（G1-18/G1-19） | T-10 |
    | V-21 | OnStart/OnReady 失败与挂起、准备期硬截止、提交后进入 Degraded、ACTIVATED 之前不开放业务、10s 共享停止预算并行生效（G1-16/G1-17/G1-18） | T-07 停止集成；T-08 配置/DB 回调；T-09 正式候选的提交前后恢复 |
    | V-36 | 双 Core（同进程/跨进程/真实二进制）、按插件串行、旧实例持有 runtime.lock 时不能启动第二个写入者、Core 被 kill 后 SDK 断链退出（G1-12/G1-15/G1-19） | T-08 DB 锁与候选资源；T-09 正式 enable/update 与仍在写 DB 的子进程 |
    | V-37 | Bootstrap 与 REGISTER 超过 4 MiB 时拒绝、SESSION_ERROR 带失败阶段、没有半注册、ACTIVATED 严格先于业务帧（G1-16/G1-17） | T-08 旧配置先经 OnLoad 再验证；T-09 |
    | V-42 | 每方向 256 个/16 MiB 与全局 64 MiB 满额时立即拒绝，Response/Cancel 仍能推进，发送前取消的请求被撤回，Cancel 不越过其 Request（G1-14） | T-06 Service 嵌套与 Event Bus 双预算 |
    | V-32/V-34 | 可复用的真实进程夹具（cmd e2e、supervisor 辅助进程）；五平台构建通过（G1-11） | T-10 五平台运行与 100 次启停 |
  - **T-02「停止和失败条件」核对**：协议、身份、门禁、锁、取消、停止的用例都通过；最小进程切片可以重复运行（-count=3）；配置/DB/安装模块的接入位置清楚（Bootstrap.database_descriptor、`supervisor.Spec`、`host.ActivateOptions.Commit`、`state` 的迁移表）。没有出现以下情况：双代次同时接单、旧写入者未退出就启动替代者（并发启动测试证明了这一点）、满额时控制帧永久阻塞、丢失有效错误结果、状态库损坏后静默初始化。**结论：T-02 在 B1 范围内完成**；尚未运行的目标平台保持「未运行」。
  - 剩余风险：业务分发未实现（SDK 对业务请求一律回 UNSUPPORTED）；Degraded 与失败窗口只在内存中；cmd 只是开发入口（安装记录的来源为 dev-local，不校验签名）；Windows Job Object 与 Windows 传输**未运行**。
  - 下一步：REVIEW-6。

### [x] REVIEW-6 集中检查（G1-17～G1-19）
- 检查清单同上；另外核对 T-02「停止和失败条件」逐条满足，并用 code-review-expert 审查整个分支。
- 回写：
  - 流程说明：code-review-expert 的 SKILL.md 第 7 步要求先确认再修改；按 goal-mode 的 SKILL.md「红线」，审查发现直接登记为 FIX task。
  - 检查过程与结论（审查范围 `c55e4c8..47682ef`：G1-17～G1-19，共 15 个文件，约 2340 行）：
    1. 是否偏离 input.md：没有；直接依赖正好是已批准的 7 项（x/sys 自 G1-12 起为直接依赖），`go mod tidy` 前后没有差异，`gofmt -l` 为空。
    2. 五平台 `CGO_ENABLED=0 go build` 与 `go vet` 全部 OK；全仓库 `go test -race -count=3 ./...` 没有失败的包，结束后 `pgrep` 确认没有残留进程。
    3. **T-02「停止和失败条件」逐条核对**：结论与 G1-19 回写一致（协议、身份、门禁、锁、取消、停止用例都通过；最小进程切片可以重复运行；没有出现双代次接单、旧写入者未退出就启动替代者、控制帧永久阻塞、丢失有效错误结果、状态库损坏后静默初始化）。
    4. code-review-expert（COMMENT，没有 P0/P1）：
       - **P2-1 `supervisor.Running.Stop` 没有持有逐插件管理锁**：只有 `Launch` 在 `store.LockPlugin` 内执行，停用（disable）可能与同一插件的另一次启动交错，违反 CONTRACTS §12.2「每插件的 install/enable/disable/… 操作串行化」。修复：Running 保存 store 与 pluginID，Stop 时获取同一把锁。
       - **P3-1 `supervisor.StopAll` 没有调用方，也没有测试**（并行停止已由经过测试的 `host.StopAll` 提供）→ 删除。
       - **P3-2 README.md 的状态过时**（仍写着「尚未创建正式 Core、SDK 或插件实现」）→ 更新为 B1 完成后的实际状态与入口。
       - 记录但不修：`Instance.Stop` 会把 Degraded 状态覆盖为 Draining，`WaitExit` 不把异常退出标记为 Degraded/Crashed——只影响内存中的诊断状态，T-09 的恢复编排会一起重做状态持久化；本机其他进程可以抢先连上 Core 的 loopback Listener，后果只是认证失败、启动失败（拒绝服务，不会越权），属于 CONTRACTS §7.1 已声明的同主机信任边界。
    5. 安全：认证、Secret 不进入日志、失败路径安全消息、精简环境、锁文件权限都由测试覆盖；cmd 的开发安装记录来源为 dev-local，不校验签名（预期内，由 T-09 实现）。
  - 下一步：FIX-6，然后终审。

---

## 终审（全部 task 完成后）

- 对照 input.md 与 plan.md §1 检查完成定义；最大范围复查（代码质量、并发、安全、错误处理、测试覆盖、交叉编译、文档、回滚方案）；汇总所有阻塞项与「未运行」项；在本节写出结论。
- 结论（2026-10-21，goal-1 分支，终审时的 HEAD 为 9c9f753 之上的本次提交）：**goal-1（B1 = T-01 + T-02）完成，没有阻塞项。**
  - 后续说明：终审之后的独立复查用临时探针复现了七项行为缺陷，这个结论因此不再成立；缺陷与修复见文末「B1 修复与规范落地」。本条作为历史记录保留。
  - **对照 input.md**：按要求先用 init-project 初始化（main 分支 5f9e903），再用 goal-mode 持续推进；只做 B1；依赖只使用已批准清单（直接依赖 7 项）；PG 使用临时容器、用完即删；Git 只在 goal-1 分支（worktree `.worktrees/goal-1`）上逐 task 提交（共 32 个 commit），**没有 push，也没有合并到 main**。组织名 `AlonaBot` 无法通过 API 注册（REST 返回 404），需要用户在网页上注册。
  - **对照 plan.md §1 完成定义**：T-01 —— proto 可以编译且生成稳定（本次终审重新执行 `go generate` 与 golden `-update` 后均无 diff）、Envelope 字段号与枚举值由测试锁定、manifest/release/消息结构都有校验与 fixture、SDK 公共类型与示例可以编译、SQLite（CRUD/Tx/WAL Backup/账本同事务）与 PG（临时容器）都通过、五平台 `CGO_ENABLED=0` 构建通过、版本写入 go.mod/go.sum（`go mod verify` 显示 all modules verified）✓。T-02 —— 真实 SDK 插件在 TCP 与 stdio 上都走完 Hello→…→STOPPING→退出；core.lock、runtime.lock、调用表、准入、取消顺序、阶段预算、停止预算、有限重启都有测试；V 分项证据表见 G1-11（T-01）与 G1-19（T-02）✓。B1 之外的能力没有实现，只预留了接入点 ✓。
  - **最大范围复查**：全仓库 `go test -race -count=1 ./...` 的 15 个包都 ok；`go test -v ./...` 统计为 **272 PASS / 0 FAIL / 2 SKIP**（SKIP 是两个 PG 测试，未设置 `ALONABOT_TEST_PG_DSN`，已在 G1-09 的容器上实测通过）；cmd e2e 与 supervisor `-count=2` ok；五平台 build+vet OK；archtest 通过；文档相对链接全部有效；没有 TODO/FIXME/DEBUG 残留；`pgrep` 显示没有残留测试进程。代码规模：非测试 Go 代码约 5.1k 行（不含生成代码），测试约 5.0k 行。
  - **质量视角**：并发（Session/激活/停止经过多轮 `-race -count=10` 与变异检查）；安全（恒时比较 Secret、认证失败时不回复、安全消息与日志中不含 Secret/配置/GORM 参数值、精简环境、锁文件 0600、路径与 ID 校验）；数据一致性（CommitActive 核对预期旧代次、提交失败不激活、提交后失败进入 Degraded 而不回滚、状态库损坏不重建且只读预检不改写文件）；错误处理（失败路径限时通知并强杀后确认进程退出）；回滚方案（plan.md §7 仍然有效：整个分支未合并到 main，可以整体丢弃；PG 容器已清理）。
  - **复查中发现并修复的问题汇总**（FIX-1～FIX-6 以及各 task 内发现的问题）：服务调用缺少可信上下文载荷；Chain 校验的内存放大（296 MiB→2.5 KB）；GORM 日志泄漏参数值；HTTP 路径未规范化；相对路径 data_dir 无法重新打开；控制帧构造重复与 stdio 管道关闭；停用未持有插件锁；失败路径的写操作在对端不读时卡死；Session 回调引用未赋值的指针（数据竞争）；Admit 看不到 parent_call_id；archtest 在取不到 VCS 状态的环境下失败；测试本身的问题（空转的进程组测试、偶发的满额测试、race atexit sleep）。
  - **「未运行」与已知限制（交给后续批次）**：Linux/Windows **运行**都未进行（只有交叉构建与 vet；darwin/amd64 只经 Rosetta 跑过单元测试）——T-10；Windows Job Object、LockFileEx、只读 URI 打开、stdio 写截止都未运行；PG 只在 G1-09 实测过一次；没有 CI；Degraded 与失败窗口只在内存中——T-09；三处 `shortcut:`（临时 route_id → T-04；SDK 业务分发返回 UNSUPPORTED → T-03/T-04；Windows 进程在 Start 之后才加入 Job）；cmd 只是开发入口（来源 dev-local，不验签）——T-09；配置迁移/Secret/候选 DB（T-08）、Telegram（T-03）、Scheduler/HTTP（T-07）尚未开始。
  - **规格改动**：WIRE-FORMAT 新增 §5A「T-01 已固定」（proto 是编号的权威来源、T-01 的实现选择、schema_id 约定、握手失败语义、启动环境变量、PREPARED 摘要）；adr/DECISIONS.md 加了一行变更记录。没有修改既有的 Envelope 1～16 与错误码 1～17。
  - **需要用户处理**：(1) 在 GitHub 网页上注册组织 `AlonaBot`；(2) 审阅 `goal-1` 分支后自行决定是否合并到 main（本 goal 不合并、不 push）；(3) 确认后可以用 goal-2 开始 B2（T-03 + T-04）。

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

### [x] FIX-2 Chain 实体边界检查的内存放大，以及 contracts 文档状态（来自 REVIEW-2，排在 G1-07 之前）
- 依据：CONTRACTS §1.2（Core 合计 64 MiB 的 payload 预算）；REVIEW-2 第 5、6 条。
- 内容：`validate/chain.go` 中把每个字符边界的 map 改为按 entity 的偏移检查：先算出每个 rune 的 UTF-16 长度，只在需要的偏移上判断是否落在边界（例如用 `[]bool` 或排序后的偏移做一次线性扫描），保证分配量与文本长度同阶且常数很小，避免使用 `utf16.Encode` 为每个 rune 分配；在测试中加入内存回归断言（4 MiB 文本加 1 个 entity，`testing.AllocsPerRun` 或 MemStats 显示分配量明显低于 64 MiB，例如低于 16 MiB）。更新 ARCHITECTURE.md 中 contracts 行的状态（proto、生成、validate、manifest、release 已实现，SDK 相关仍是拟建）。
- 完成条件：原有 43 个 Chain 用例仍然通过；新增的内存回归测试通过；修改前后各测一次大文本的分配量并记录到这里。
- 回写：
  - 实际做了什么：`validate/chain.go` 中的边界检查改为只把实体用到的起止偏移放进 map（大小为 2×实体数），扫描一遍文本，用 `utf16.RuneLen` 累加 UTF-16 位置，并标记这些偏移是否落在边界上；不再给每个字符建 map 项，也不再为每个 rune 调用 `utf16.Encode`。`chain_test.go` 新增 `TestChainLargeTextAllocation`（约 4 MiB 的中文文本加 1 个 entity，要求 TotalAlloc 增量不超过 1 MiB）。ARCHITECTURE.md 中 contracts 的状态改为「已实现：proto/生成、validate、manifest、release」。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：分配量对比（用 MemStats 的 TotalAlloc 测量）：旧实现在 4 MiB ASCII 文本上为 296 MiB（REVIEW-2 探针），在本用例的中文文本上为 78,464,976 字节；新实现在同一中文文本上为 2,520 字节。把旧实现临时放回来时，`TestChainLargeTextAllocation` 失败（"分配了 78464976 字节"）；换回新实现后通过。`go test -race -count=1 ./contracts/...` 的 4 个包都 ok；TestChain 系列共 44 个 PASS（原有 43 个加新增的 1 个）；`gofmt -l contracts` 为空。
  - 过程中的失误（已纠正）：shell 里的 `cp` 是交互式别名，第一次把新实现拷回来时没有生效，`git checkout` 又让测试改动丢失；随后从 scratchpad 备份恢复了新实现，并重新加上了回归测试。最终 diff 只包含 chain.go 与 chain_test.go 的预期改动。
  - 剩余风险：TotalAlloc 断言会受 GC 时机与运行时版本影响，但 1 MiB 的阈值与实测约 2.5 KB 相差三个数量级，误报的可能性很低。
  - 下一步：G1-07。

### [x] FIX-3 GORM 日志不得包含参数值、HTTP 路径规范化、ARCHITECTURE 同步（来自 REVIEW-3，排在 G1-10 之前）
- 依据：CONTRACTS §5.3、§7.2；REVIEW-3 中的 P2-1、P2-2、P3-2。
- 内容：`sdk/gormdb/sqlite.go` 的 gormLogger 设置 `ParameterizedQueries: true`，并新增测试：把 logger 输出重定向到 buffer，执行一条带参数的失败 SQL，断言日志中有 SQL 文本、但没有参数值；`sdk/registry.go` 的 `HTTPRegistry.Handle` 要求路径以 / 开头且 `path.Clean(p)==p`，并在 sdk_test 中补上 `/a/../b`、`//x`、`/a/` 这几个被拒绝的用例；更新 ARCHITECTURE.md 中 sdk 行的描述。
- 完成条件：新测试通过；先撤掉 ParameterizedQueries 验证新测试会失败，再恢复；全仓库 `go test -race ./...` 通过。
- 回写：
  - 实际做了什么：`sdk/gormdb/sqlite.go` 改为通过 `newGormLogger(io.Writer)` 构造 logger，生产环境写 stderr，并设置 `ParameterizedQueries: true`；SQLite 与 PG 共用这个 logger。新增 `TestGormLogOmitsParameters`：用带 buffer 的 logger 执行一条带参数、因违反外键而失败的 INSERT，断言日志里有 `INSERT INTO child`，但没有参数值 `SECRET-VALUE-42`。`sdk/registry.go` 的 `HTTPRegistry.Handle` 要求路径以 / 开头且满足 `path.Clean(p)==p`。sdk_test 新增 `/a/../b`、`//x`、`/a/` 三个应被拒绝的用例，以及 `TestHTTPRootPathAllowed`（`/` 允许注册）。ARCHITECTURE.md 中 sdk 行补上了 OpenSQLite/OpenPostgres/RunVersioned。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：`go vet ./... && go test -race -count=1 ./...` 的 7 个包都 ok；`gofmt -l sdk` 为空。变异检查：把设置改成 `ParameterizedQueries: false` 后，测试失败，日志原文为 `INSERT INTO child(id, parent_id, note) VALUES (1, 99, 'SECRET-VALUE-42')`，证实修复前确实会泄漏参数值；恢复设置后通过。
  - 剩余风险：只覆盖 GORM 自己的日志；插件自行打印的日志里有没有业务数据，属于可信插件作者的责任（CONTRACTS §7.1）。
  - 下一步：G1-10（架构 import 守卫）。

### [x] FIX-4 状态库相对路径 data_dir 无法重开，以及 x/sys 文档（来自 REVIEW-4，排在 G1-14 之前）
- 依据：REVIEW-4 第 5、7 条；CONTRACTS §6.1（状态库错误不得变成不可启动的误报）。
- 内容：`state.Open` 一开始就把 dataDir 转成绝对路径（`filepath.Abs`），后续的锁文件、状态库与预检 URI 都使用这个绝对路径；新增测试：用 `t.Chdir` 切换到临时目录，以相对路径连续 Open/Close 两次都成功，且第二次读到第一次写入的记录。修正 docs/development.md 中关于 x/sys 的描述。
- 完成条件：新测试在修复前失败、修复后通过（先用旧实现验证一次失败）；全仓库 `go test -race ./...` 通过；Windows 交叉 vet 通过。
- 回写：
  - 实际做了什么：`state.Open` 一开始就执行 `dataDir, err := filepath.Abs(dataDir)`，之后的锁文件、state.sqlite 与预检 URI 都基于这个绝对路径；preflight 的注释写明路径必须已经是绝对路径，「前面补 /」只用于 Windows 盘符。新增 `TestRelativeDataDirReopens`：用 `t.Chdir(t.TempDir())` 切换目录后，以相对路径 `reldata` 连续 Open/Close 两次，第一次写入的安装记录在第二次能读到。docs/development.md 改为「x/sys v0.49.0 为直接依赖，用于 internal/oslock」。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：**修复前**新测试失败：「第 2 次以相对路径打开失败: 状态库无法读取（可能已损坏，不会重建）: unable to open database file (14)」，与 REVIEW-4 的实测一致；**修复后**通过。全仓库 `go test -race -count=1 ./...` 的 11 个包都 ok；`GOOS=windows go vet ./internal/...` 通过；`gofmt -l internal` 为空。
  - 剩余风险：Windows 的绝对路径转成 URI（`file:///C:/...`）只经过 vet，没有实际运行。
  - 下一步：G1-14（Session 调用引擎）。

### [x] FIX-5 合并控制帧构造、Release 关闭 stdio 管道、缺少阶段预算时报协议错误（来自 REVIEW-5，排在 G1-17 之前）
- 依据：REVIEW-5 中的 P3-1/P3-2/P3-3。
- 内容：`internal/wire` 新增 `ControlEnvelope`，`Session.SendControl`、`host`、`sdk` 统一改用它，删除各自的副本；`process.Release` 显式关闭 Stdin/Stdout（stdio 模式）；`sdk.prepare` 在 PhaseLimits 或 preparation_deadline 缺失、OnStart 预算为 0 时发送 SESSION_ERROR(CONFIG, PROTOCOL_ERROR)，并补一个测试。
- 完成条件：全仓库 `go test -race ./...` 通过；新测试覆盖缺少预算的情况；三平台 vet 通过；archtest 通过。
- 回写：
  - 实际做了什么：`internal/wire` 新增 `ControlEnvelope(sid, minor, sourcePluginID, kind, msg)`（PROTOBUF 编码、确定性序列化、schema_id 为 `alonabot.<snake>.v1`）。共有五处改用它：`Session.SendControl`、`Session.sendCancel`（新增私有方法 `Session.controlEnvelope` 填入本端的 session/minor/source）、`host` 的 HELLO_ACK/PREPARED/SESSION_ERROR（删除 `host.controlEnvelope`）、`sdk` 的 HELLO/REGISTER/SESSION_ERROR（删除 `sdk.controlEnv`）。顺带修正：SDK 握手之后发出的帧，minor 改用协商得到的 `ack.ChosenMinor`，原来固定为 0。`process.Release` 现在会显式关闭 stdio 模式下 Core 持有的 Stdin/Stdout，重复调用不报错，并把 os.ErrClosed 视为已经关闭。`sdk.prepare` 在 PhaseLimits 缺失、OnStart 预算为 0 或 preparation_deadline 缺失时，发送 SESSION_ERROR(CONFIG, PROTOCOL_ERROR)，并且不执行 OnStart。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：新增 `TestPrepareRejectsMissingPhaseLimits`（收到 CONFIG/PROTOCOL_ERROR，OnStart 没有被调用）；`TestStdioTransportPipes` 新增断言：Release 之后向 Stdin 写入失败，重复 Release 不报错。全仓库 `go test -race -count=1 ./...` 的 13 个包都 ok（其中 golden 帧与 host/sdk 握手测试不受影响）；`GOOS=windows/linux go vet ./...` 通过；`grep` 确认 `controlEnv`/`controlEnvelope(sid` 的副本已经删除；archtest 通过。
  - 剩余风险：无新增风险。
  - 下一步：G1-17（提交与激活门禁）。

### [x] FIX-6 停用持有逐插件锁、删除未用的 StopAll、README 状态（来自 REVIEW-6，排在终审之前）
- 依据：REVIEW-6 中的 P2-1/P3-1/P3-2；CONTRACTS §12.2。
- 内容：`supervisor.Running` 记录 store 与 pluginID，`Stop` 在 `store.LockPlugin(pluginID)` 内执行；新增测试：插件 Stop 正在进行时，同一插件的 Launch 要等 Stop 完成后才开始（用会挂起 STOPPING 的辅助进程拉长 Stop）；删除 `supervisor.StopAll`；README.md 更新为 B1 的实际状态（T-01/T-02 已完成、开发入口、文档链接）。
- 完成条件：新测试通过（并验证在去掉锁之后会失败）；全仓库 `go test -race ./...` 与 cmd e2e 通过；三平台 vet 通过。
- 回写：
  - 实际做了什么：`supervisor.Running` 增加 `store`、`pluginID` 两个字段，`Stop` 在 `store.LockPlugin(pluginID)` 内执行，与 `Launch` 共用逐插件管理锁；删除没有调用方的 `supervisor.StopAll`（并行停止由经过测试的 `host.StopAll` 提供）；新增 `TestStopAndLaunchSerializedPerPlugin`：用不响应 STOPPING 的辅助进程把 Stop 拉长到总截止 800ms，在 Stop 进行中对同一插件发起 Launch，断言 Launch 在 Stop 完成之后才返回。README.md 改为 B1 之后的实际状态（T-01/T-02 已完成，只有开发入口 `alonabot run`，并给出文档链接）。
  - 验证证据（darwin/arm64，go1.27.1，本 task 的 commit）：新测试 PASS（0.85s）；**变异检查**：去掉 Stop 中的锁之后，测试报「同一插件的 Launch 在 Stop 完成（13:53:32.252）之前就返回（13:53:31.453），管理操作未串行」，恢复后通过；全仓库 `go test -race -count=1 ./...` 没有失败的包；`go test -count=2 ./cmd/alonabot/` ok；三平台 vet 通过；`pgrep` 确认没有残留进程。
  - 剩余风险：没有新增风险。
  - 下一步：终审。

## B1 修复与规范落地（codex/b1-repair）

终审之后的复查发现七项缺陷：原有测试全部通过，但临时探针（`tmp/test/_probe/b1-review`，不提交）能稳定复现。修复在 `codex/b1-repair` 分支上进行（从 goal-1 的 49fcb29 分出，R-00 合并了 main 的 b1559cb：项目规范、仓库技能与 CI），goal-1 的历史保持不变。行为修复、纯移动和注释调整分开提交。

### [x] R-00 建立整合基线
- 内容：建立 worktree `.worktrees/b1-repair`，合并 main。唯一的冲突在 docs/development.md「规则与检查」表，处理时保留 B1 已观测的四项状态，再追加 main 新增的两行（3514f43）。七项缺陷在整合后的分支上用原探针全部复现。

### [x] R-01 取锁顺序与相对 data_dir（d72a8f2）
- 缺陷 1：`sdk.Run` 先调用 `Collect`（会执行用户的 Commands 等注册入口），后取 runtime.lock。旧实例仍在时，用户代码已经执行。
- 缺陷 2：相对 `--data-dir` 下，`ALONABOT_RUNTIME_LOCK` 是相对路径，SDK 以「必须是绝对路径」拒绝，插件无法启动。
- 修复：`Run` 先取锁再 `Collect`，`Collect` 本身不加锁前置条件；`supervisor.Launch` 把 PluginDir 绝对化一次，工作目录与锁路径共用这个结果。
- 回归：`TestRunTakesLockBeforeRegistration`，`TestRelativeDataDir`（TCP/stdio，各跑 new + restart，锁文件保留）。

### [x] R-02 进程树回收与退出分类（edc88a0、9e37561）
- 缺陷 3：领头进程退出后，留在进程组里的子进程不会被清理，Stop 和崩溃恢复都直接放行。
- 缺陷 4：退出码 0 被判为异常退出，进入重启计数，可能导致插件被禁用。
- 修复：
  - `process.KillTree` 终止后，确认整棵树都已消失才返回；新增 `WaitTree`。Unix 用 `kill(-pgid, 0)` 判断，Windows 查询 Job 的 ActiveProcesses。
  - `host.Stop` 在领头进程退出后，用同一个总截止等待剩余进程，到期后终止，错误放进 `StopResult.Err`。
  - `supervisor` 用 mutex 决定 Stop 与 WaitExit 由谁清理，资源只释放一次。未经 Stop 的退出由 WaitExit 清理剩余进程树，失败时返回错误。只有「未经 Stop 且退出码非 0」才算异常。
  - `Launch` 失败路径保留 KillTree 的错误（9e37561）。
  - cmd 在进程树未确认清空时报错退出，不启动替代实例。
- 回归：`TestNormalExitIsNotAbnormal`，`TestLeftoverChildReapedBeforeReturn`（stop 与 crash 两条路径）。变异验证：分别去掉 WaitTree、WaitExit 中的 KillTree、Success 判断，对应测试都会失败。
- 未运行：Windows 的 Job 清空确认只做了交叉 vet。

### [x] R-03 取消请求后的队列与预算（02c8d1a）
- 缺陷 5：取消排队中的请求时，预算立即归还，但 `s.biz` 仍持有载荷。writer 阻塞时反复排队再取消，就能越过字节上限（探针：队列保留 3000 字节，上限 1000）。
- 修复：调用状态统一由 Session.mu 保护，`complete` 按写出状态分别处理：
  - 排队中：移出队列、丢弃帧并归还预算，对端既收不到 Request 也收不到 Cancel。
  - 写出中：写完再归还预算，然后按放弃原因发 Cancel。
  - 已写出：立即归还预算并发 Cancel。
  - 预算只归还一次；写完后补发的 Cancel 使用实际原因，原来固定写 CALLER_CANCELLED。
- 回归：`TestCancelledRequestsReleaseQueueAndBudget`，用旧实现跑会失败。

### [x] R-04 认证后帧校验与错误码映射（bc66e83）
- 缺陷 6：握手后没有持续核对帧的 source_plugin_id 和协商 minor，插件可以冒用其他 ID，Core 帧也可以伪带插件来源。
- 缺陷 7：RESPONSE 中的未知错误码原样交给调用者。
- 修复：
  - 新增 `wire.CheckPeer`，由 Session 读循环、Core 准备阶段（REGISTER/SESSION_ERROR）、SDK 的 HELLO_ACK/PREPARED 共用；Core 侧 Session 传入插件 ID。
  - `responseOf` 用 `validate.PublicCode` 公开错误码，保留载荷、safe_message 和 trace_id，原始码放进 `Response.RawErrorCode`；SDK 收到的 SESSION_ERROR 同样做映射。
  - 修正了两个不合 WIRE 规定的测试夹具。
- 回归：`TestSessionRejectsForeignIdentityOrMinor`、`TestPrepareRejectsForeignSourceOrMinor`、`TestUnknownErrorCodeMappedToInternal`，均做了变异验证。

### [x] R-05 注释补齐与最小结构整理（7400c0f～e802828）
- 结构：只把 `Budget` 及其方法原样移到 `internal/wire/budget.go`；`activate_test` 中重复三次的匿名结果结构改为具名类型 `activation`，没有其他搬迁。
- 注释：
  - 所有手写 Go 文件（含测试、示例）、9 份 .proto 和 YAML fixture 都有职责头。
  - 类型、字段（含私有、嵌入、局部和匿名结构体字段）、接口方法和具名函数都有说明，合并声明的字段已拆开。
  - 多阶段函数写了按真实阶段编号的 step comment。
  - proto 的说明写在源 Schema 中，重新生成后已核对，生成文件只有注释和空白差异。
  - frames/*.bin 与 release.json 不能写注释，用途在 contracts/doc.go 中说明。
- 检查：临时 AST 检查器（`tmp/test/_probe/commentcheck`，不提交）最终报告 0 处缺口；它只检查是否存在说明，说明是否准确靠人工核对。

### [x] R-06 文档与验收记录
- 更新内容：
  - spec/TEST-MATRIX.md 新增 §3.1，按 V 编号记录 T-01/T-02 分项的测试位置、命令、平台、结果和承接任务，没有关闭任何整项 V。
  - spec/PLAN.md 与 spec/README.md 更新实施状态。
  - WIRE-FORMAT §6 和 PLAN §2 中「尚未执行」的陈述已更正。
  - ARCHITECTURE.md 的当前状态、wire 与 sdk 两行已更正。
  - docs/development.md 的 V 证据状态已更新。
  - README.md 写明只在 darwin/arm64 运行过。

### R-07 门禁证据（候选 e802828，darwin/arm64，Go 1.27.1）
- gofmt 无输出；`go generate ./...` 后无差异；`go vet ./...` 通过，linux/amd64 与 windows/amd64 的 vet 也通过；五目标 `CGO_ENABLED=0 go build ./...` 都通过；`go mod verify` 显示 all modules verified；`node --test scripts/check-repository.test.mjs`（2 pass）、`node scripts/check-repository.mjs` 和 `git diff --check` 都通过。
- 全仓库测试：设置 `ALONABOT_TEST_PG_DSN` 后执行 `go test -race -count=1 ./...`，15 个包、291 个测试（含子测试）全部 PASS，0 FAIL，0 SKIP。PG 用一次性 postgres:17 容器（a63c1ea3f6a3），跑完确认没有残留 `ab_%` role/schema，随后删除容器。另跑 `go test -race -count=3 ./cmd/alonabot/ ./internal/core/supervisor/`，通过。
- 未运行：Linux 与 Windows 的真实运行（T-10）；Windows Job 清空确认与 LockFileEx；PR head 的 CI 结果在 PR 中记录。
