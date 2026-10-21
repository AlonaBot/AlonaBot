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
