package i18n

// messages 包含用户可见的错误消息（存储在API响应的error字段）
// 键格式: entity.operation (点分，小写)
var messages = map[string]map[string]string{
	// --- 通用错误 ---
	"unauthorized": {
		"en":    "Unauthorized",
		"zh-CN": "未授权",
	},
	"forbidden": {
		"en":    "Forbidden",
		"zh-CN": "禁止访问",
	},
	"not_found": {
		"en":    "Not found",
		"zh-CN": "未找到",
	},
	"rate_limited": {
		"en":    "Too many requests, please slow down",
		"zh-CN": "请求过于频繁，请稍后再试",
	},
	"rate_limited_admin_login": {
		"en":    "Too many login attempts, please try again later",
		"zh-CN": "登录尝试次数过多，请稍后再试",
	},
	"validation_error": {
		"en":    "Validation failed: %s",
		"zh-CN": "参数校验失败: %s",
	},
	"time.invalid_filter": {
		"en":    "%s: %s",
		"zh-CN": "%s: %s",
	},

	// --- 任务 ---
	"task.not_found": {
		"en":    "Task not found",
		"zh-CN": "未找到任务",
	},
	"task.invalid_id": {
		"en":    "Invalid task ID",
		"zh-CN": "无效的任务ID",
	},
	"task.invalid_status_filter": {
		"en":    "Invalid status filter",
		"zh-CN": "无效的状态过滤器",
	},

	// --- 认证 ---
	"auth.invalid_credentials": {
		"en":    "Invalid username or password",
		"zh-CN": "用户名或密码错误",
	},
	"auth.username_taken": {
		"en":    "Username already taken",
		"zh-CN": "用户名已被占用",
	},
	"auth.invalid_old_password": {
		"en":    "Invalid old password",
		"zh-CN": "原密码错误",
	},
	"auth.missing_api_key": {
		"en":    "Missing API key (use Authorization: Bearer <key> or api-key: <key>)",
		"zh-CN": "缺少API密钥（使用 Authorization: Bearer <key> 或 api-key: <key>）",
	},
	"auth.invalid_api_key": {
		"en":    "Invalid API key",
		"zh-CN": "无效的API密钥",
	},

	// --- 标签 ---
	"tag.not_found": {
		"en":    "Tag not found",
		"zh-CN": "未找到标签",
	},
	"tag.invalid_id": {
		"en":    "Invalid tag ID",
		"zh-CN": "无效的标签ID",
	},
	"tag.name_taken": {
		"en":    "Tag name already taken",
		"zh-CN": "标签名已被占用",
	},

	// --- 提醒配置 ---
	"config.not_found": {
		"en":    "Config not found",
		"zh-CN": "未找到配置",
	},
	"config.invalid_id": {
		"en":    "Invalid config ID",
		"zh-CN": "无效的配置ID",
	},

	// --- 用户 ---
	"user.not_found": {
		"en":    "User not found",
		"zh-CN": "未找到用户",
	},
	"user.invalid_id": {
		"en":    "Invalid user ID",
		"zh-CN": "无效的用户ID",
	},

	// --- API Key ---
	"apikey.not_found": {
		"en":    "API key not found",
		"zh-CN": "未找到API密钥",
	},
	"apikey.invalid_id": {
		"en":    "Invalid key ID",
		"zh-CN": "无效的密钥ID",
	},

	// --- 管理后台 ---
	"admin.access_required": {
		"en":    "Admin access required",
		"zh-CN": "需要管理员权限",
	},
	"admin.cannot_remove_own_privilege": {
		"en":    "Cannot remove your own admin privileges",
		"zh-CN": "不能移除自己的管理员权限",
	},
	"admin.localhost_only": {
		"en":    "Admin interface is localhost-only",
		"zh-CN": "管理后台仅限本地访问",
	},

	// --- 系统日志 ---
	"system.log_file_not_found": {
		"en":    "Log file not found",
		"zh-CN": "日志文件未找到",
	},
	"system.invalid_log_filename": {
		"en":    "Invalid log filename",
		"zh-CN": "无效的日志文件名",
	},
	"system.invalid_status_filter": {
		"en":    "Invalid status, must be 'success' or 'failed'",
		"zh-CN": "无效的状态值，必须为 'success' 或 'failed'",
	},
	"system.endpoint_not_found": {
		"en":    "Endpoint not found",
		"zh-CN": "接口不存在",
	},

	// --- 用户资料 ---
	"profile.no_fields": {
		"en":    "No fields to update",
		"zh-CN": "没有需要更新的字段",
	},
	"profile.not_found": {
		"en":    "User profile not found",
		"zh-CN": "未找到用户资料",
	},
	"profile.email_code_required": {
		"en":    "Verification code is required to change email",
		"zh-CN": "更换邮箱需要验证码",
	},
	"profile.email_taken": {
		"en":    "Email is already in use by another account",
		"zh-CN": "该邮箱已被其他账号使用",
	},
	"profile.password_already_set": {
		"en":    "Password is already set, use change password instead",
		"zh-CN": "已设置过密码，请使用修改密码功能",
	},

	// --- OAuth ---
	"oauth.unlink_last_method": {
		"en":    "Cannot unlink: you must keep at least one login method",
		"zh-CN": "无法解绑：至少需要保留一种登录方式",
	},
	"oauth.already_linked": {
		"en":    "This account is already linked",
		"zh-CN": "该账号已经绑定过了",
	},
	"oauth.linked_to_other": {
		"en":    "This account is linked to another user",
		"zh-CN": "该账号已绑定到其他用户",
	},
	"oauth.account_not_found": {
		"en":    "OAuth account not found",
		"zh-CN": "未找到 OAuth 账号",
	},

	// --- 邮箱验证 ---
	"email.not_configured": {
		"en":    "Email service is not configured",
		"zh-CN": "邮箱服务未配置",
	},
	"email.required": {
		"en":    "Email is required",
		"zh-CN": "邮箱为必填项",
	},
	"email.code_required": {
		"en":    "Verification code is required",
		"zh-CN": "请输入验证码",
	},
	"email.code_cooldown": {
		"en":    "Please wait 60 seconds before requesting another code",
		"zh-CN": "请等待60秒后再请求验证码",
	},
	"email.code_invalid": {
		"en":    "Invalid verification code",
		"zh-CN": "验证码错误",
	},
	"email.code_expired": {
		"en":    "Verification code has expired",
		"zh-CN": "验证码已过期",
	},
	"email.code_attempts_exceeded": {
		"en":    "Too many verification attempts, please request a new code",
		"zh-CN": "验证次数过多，请重新获取验证码",
	},
	"email.code_not_found": {
		"en":    "No valid verification code found, please request a new one",
		"zh-CN": "未找到有效验证码，请重新获取",
	},
	"auth.email_taken": {
		"en":    "Email is already registered",
		"zh-CN": "该邮箱已被注册",
	},
}

// internalMessages 包含技术错误消息（存储在日志和API响应的error字段）
// 键格式: entity.operation (点分，小写)
var internalMessages = map[string]map[string]string{
	// --- 任务 ---
	"task.create": {
		"en":    "Failed to create task",
		"zh-CN": "创建任务失败",
	},
	"task.get": {
		"en":    "Failed to get task",
		"zh-CN": "获取任务失败",
	},
	"task.update": {
		"en":    "Failed to update task",
		"zh-CN": "更新任务失败",
	},
	"task.delete": {
		"en":    "Failed to delete task",
		"zh-CN": "删除任务失败",
	},
	"task.toggle": {
		"en":    "Failed to toggle task",
		"zh-CN": "切换任务状态失败",
	},
	"task.list": {
		"en":    "Failed to list tasks",
		"zh-CN": "获取任务列表失败",
	},

	// --- 标签 ---
	"tag.create": {
		"en":    "Failed to create tag",
		"zh-CN": "创建标签失败",
	},
	"tag.get": {
		"en":    "Failed to get tag",
		"zh-CN": "获取标签失败",
	},
	"tag.update": {
		"en":    "Failed to update tag",
		"zh-CN": "更新标签失败",
	},
	"tag.delete": {
		"en":    "Failed to delete tag",
		"zh-CN": "删除标签失败",
	},
	"tag.list": {
		"en":    "Failed to list tags",
		"zh-CN": "获取标签列表失败",
	},

	// --- 提醒配置 ---
	"config.create": {
		"en":    "Failed to create config",
		"zh-CN": "创建配置失败",
	},
	"config.get": {
		"en":    "Failed to get config",
		"zh-CN": "获取配置失败",
	},
	"config.update": {
		"en":    "Failed to update config",
		"zh-CN": "更新配置失败",
	},
	"config.delete": {
		"en":    "Failed to delete config",
		"zh-CN": "删除配置失败",
	},
	"config.list": {
		"en":    "Failed to list configs",
		"zh-CN": "获取配置列表失败",
	},

	// --- 认证 ---
	"auth.register": {
		"en":    "Failed to register",
		"zh-CN": "注册失败",
	},
	"auth.authenticate": {
		"en":    "Failed to authenticate",
		"zh-CN": "认证失败",
	},
	"auth.generate_key": {
		"en":    "Failed to generate key",
		"zh-CN": "生成密钥失败",
	},
	"auth.revoke_key": {
		"en":    "Failed to revoke key",
		"zh-CN": "撤销密钥失败",
	},
	"auth.list_keys": {
		"en":    "Failed to list keys",
		"zh-CN": "获取密钥列表失败",
	},
	"auth.generate_session": {
		"en":    "Failed to generate session key",
		"zh-CN": "生成会话密钥失败",
	},

	// --- 用户 ---
	"user.update_profile": {
		"en":    "Failed to update profile",
		"zh-CN": "更新用户资料失败",
	},
	"user.change_password": {
		"en":    "Failed to change password",
		"zh-CN": "修改密码失败",
	},
	"user.get": {
		"en":    "Failed to get user",
		"zh-CN": "获取用户信息失败",
	},
	"user.delete": {
		"en":    "Failed to delete user",
		"zh-CN": "删除用户失败",
	},
	"user.hash_password": {
		"en":    "Failed to hash password",
		"zh-CN": "密码哈希失败",
	},
	"user.reset_password": {
		"en":    "Failed to reset password",
		"zh-CN": "重置密码失败",
	},
	"user.update_admin": {
		"en":    "Failed to update admin status",
		"zh-CN": "更新管理员状态失败",
	},
	"user.count_tasks": {
		"en":    "Failed to count tasks",
		"zh-CN": "统计任务数失败",
	},
	"user.count_api_keys": {
		"en":    "Failed to count API keys",
		"zh-CN": "统计API密钥数失败",
	},
	"user.list": {
		"en":    "Failed to list users",
		"zh-CN": "获取用户列表失败",
	},

	// --- 提醒配置操作 ---
	"reminder_config.toggle": {
		"en":    "Failed to toggle reminder config",
		"zh-CN": "切换提醒配置失败",
	},
	"reminder_config.delete": {
		"en":    "Failed to delete reminder config",
		"zh-CN": "删除提醒配置失败",
	},
	"reminder_config.list": {
		"en":    "Failed to list reminder configs",
		"zh-CN": "获取提醒配置列表失败",
	},
	"reminder_config.create": {
		"en":    "Failed to create reminder config",
		"zh-CN": "创建提醒配置失败",
	},
	"reminder_config.update": {
		"en":    "Failed to update reminder config",
		"zh-CN": "更新提醒配置失败",
	},
	"reminder_config.get": {
		"en":    "Failed to get reminder config",
		"zh-CN": "获取提醒配置失败",
	},

	// --- 提醒日志 ---
	"reminder_log.list": {
		"en":    "Failed to list reminder logs",
		"zh-CN": "获取提醒日志失败",
	},

	// --- 邮箱服务 ---
	"email.send_code": {
		"en":    "Failed to send verification code",
		"zh-CN": "发送验证码失败",
	},
	"email.verify_code": {
		"en":    "Failed to verify code",
		"zh-CN": "验证验证码失败",
	},
	"email.test_connection": {
		"en":    "Failed to test email connection",
		"zh-CN": "测试邮箱连接失败",
	},

	// --- OAuth ---
	"oauth.link": {
		"en":    "Failed to link OAuth account",
		"zh-CN": "绑定 OAuth 账号失败",
	},
	"oauth.unlink": {
		"en":    "Failed to unlink OAuth account",
		"zh-CN": "解绑 OAuth 账号失败",
	},
	"oauth.list": {
		"en":    "Failed to list OAuth accounts",
		"zh-CN": "获取 OAuth 账号列表失败",
	},

	// --- 用户资料 ---
	"profile.get": {
		"en":    "Failed to get profile",
		"zh-CN": "获取用户资料失败",
	},

	// --- 系统统计 ---
	"system.stats": {
		"en":    "Failed to get stats",
		"zh-CN": "获取统计数据失败",
	},
	"system.task_trends": {
		"en":    "Failed to get task trends",
		"zh-CN": "获取任务趋势失败",
	},
	"system.reminder_status": {
		"en":    "Failed to get reminder status",
		"zh-CN": "获取提醒状态失败",
	},
	"system.log_list": {
		"en":    "Failed to list log files",
		"zh-CN": "获取日志文件列表失败",
	},
	"system.log_read": {
		"en":    "Failed to read log file",
		"zh-CN": "读取日志文件失败",
	},
	"system.serialize": {
		"en":    "Failed to serialize config",
		"zh-CN": "序列化配置失败",
	},
	"system.deserialize": {
		"en":    "Failed to deserialize config",
		"zh-CN": "反序列化配置失败",
	},
}
