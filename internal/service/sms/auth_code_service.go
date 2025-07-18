package sms

import (
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"haven_camp_server/internal/config"
	"haven_camp_server/internal/service/redis"
	"haven_camp_server/pkg/constants"
	"haven_camp_server/pkg/util/random"
	"haven_camp_server/pkg/zlog"
	"strconv"
	"time"
)

var smsClient *dysmsapi20170525.Client

// createClient 使用AK&SK初始化账号Client
func createClient() (result *dysmsapi20170525.Client, err error) {
	// 工程代码泄露可能会导致 AccessKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考。
	// 建议使用更安全的 STS 方式，更多鉴权访问方式请参见：https://help.aliyun.com/document_detail/378661.html。
	accessKeyID := config.GetConfig().AccessKeyID
	accessKeySecret := config.GetConfig().AccessKeySecret
	if smsClient == nil {
		config := &openapi.Config{
			// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID。
			AccessKeyId: tea.String(accessKeyID),
			// 必填，请确保代码运行环境设置了环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET。
			AccessKeySecret: tea.String(accessKeySecret),
		}
		// Endpoint 请参考 https://api.aliyun.com/product/Dysmsapi
		config.Endpoint = tea.String("dysmsapi.aliyuncs.com")
		smsClient, err = dysmsapi20170525.NewClient(config)
	}
	return smsClient, err
}

// VerificationCode 函数用于生成并发送短信验证码
// 参数：telephone - 接收验证码的手机号
// 返回值：message - 操作结果消息，code - 状态码
func VerificationCode(telephone string) (string, int) {
	// 创建Redis客户端
	client, err := createClient()
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1 // 系统错误
	}
	
	// 生成Redis存储的键名，格式为 "auth_code_手机号"
	key := "auth_code_" + telephone
	
	// 检查Redis中是否已存在该手机号的验证码
	code, err := redis.GetKey(key)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1 // 系统错误
	}

	// 如果验证码存在（未过期），直接返回提示信息
	if code != "" {
		message := "目前还不能发送验证码，请输入已发送的验证码"
		zlog.Info(message)
		return message, -2 // 验证码未过期
	}
	
	// 验证码已过期，重新生成6位随机数作为验证码
	code = strconv.Itoa(random.GetRandomInt(6))
	fmt.Println(code) // 注意：生产环境建议移除该打印，避免泄露验证码
	
	// 将新生成的验证码存入Redis，有效期1分钟
	err = redis.SetKeyEx(key, code, time.Minute)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1 // 系统错误
	}
	
	// 构建短信发送请求
	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      tea.String("阿里云短信测试"),
		TemplateCode:  tea.String("SMS_154950909"), // 短信模板
		PhoneNumbers:  tea.String(telephone),
		TemplateParam: tea.String("{\"code\":\"" + code + "\"}"),
	}

	// 设置运行时选项
	runtime := &util.RuntimeOptions{}
	
	// 调用阿里云短信服务发送验证码
	rsp, err := client.SendSmsWithOptions(sendSmsRequest, runtime)
	if err != nil {
		zlog.Error(err.Error())
		return constants.SYSTEM_ERROR, -1 // 系统错误
	}
	
	// 记录发送结果并返回成功信息
	zlog.Info(*util.ToJSONString(rsp))
	return "验证码发送成功，请及时在对应电话查收短信", 0 // 成功
}
