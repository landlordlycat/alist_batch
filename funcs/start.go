package funcs

import (
	"encoding/json"
	"log"

	"github.com/yzbtdiy/alist_batch/models"
)

var conf *models.Config

func Run() {
	//读取配置文件
	conf = GetConfig("config.yaml")

	// 检查配置文件和分享链接文件是否存在
	hasConf := CheckFile("config.yaml")
	hasAliShare := CheckFile("ali_share.yaml")
	hasPikShare := CheckFile("pik_share.yaml")

	if hasConf {
		if conf.Aliyun.Enable {
			if hasAliShare {
				Start()
			} else {
				log.Println("ali_share.yaml文件不存在, 尝试生成")
				GenAliShareFile("ali_share.yaml")
			}
		}
		if conf.PikPak.Enable {
			if hasPikShare {
				Start()
			} else {
				log.Println("pik_share.yaml文件不存在, 尝试生成")
				GenPikShareFile("pik_share.yaml")
			}
		}
	} else {
		log.Println("config.yaml文件不存在, 尝试生成")
		GenConfFile("config.yaml")
	}
}

func Start() {
	// 检查配置文件是否修改必要字段
	CheckConf(conf)

	// 拼接API
	loginApi := conf.Url + "/api/auth/login"
	storageListApi := conf.Url + "/api/admin/storage/list"
	addStorageApi := conf.Url + "/api/admin/storage/create"
	delStorageApi := conf.Url + "/api/admin/storage/delete"

	// 检测 token 是否存在
	if conf.Token != "ALIST_TOKEN" && conf.Token != "" {
		// 携带 token 尝试读取storagelist, 若返回 200 则说明 token 有效
		storageListRes := HttpGet(storageListApi, conf.Token)
		if storageListRes.Code == 200 {
			// 如果有 -delete flag, 进行删除操作
			DeleteStorageIfHaveFlag(storageListApi, delStorageApi, conf.Token)
			// 发送请求添加阿里云分享链接
			if conf.Aliyun.Enable {
				PushAliShares(addStorageApi, conf)
			}
			if conf.PikPak.Enable {
				PushPikPakShares(addStorageApi, conf)

			}
		} else {
			// 若携带 token 尝试访问 storage list 失败, 则尝试更新 token
			UpdateToken(loginApi, conf)
		}
	} else {
		// 若 token 为 ALIST_TOKEN 或空字符串, 则尝试更新 token
		UpdateToken(loginApi, conf)
	}

}

// 检查配置文件是否修改
func CheckConf(conf *models.Config) {
	if conf.Url == "ALIST_URL" {
		panic("url 未配置, 请检查配置文件")
	}
	if (conf.Auth.Username == "USERNAME" || conf.Auth.Password == "PASSWORD") && (conf.Token == "ALIST_TOKEN" || conf.Token == "") {
		panic("token和用户密码至少要配置一项, 请检查配置文件")
	}
	if conf.Aliyun.Enable && conf.Aliyun.RefreshToken == "ALI_YUNPAN_REFRESH_TOKEN" {
		panic("添加阿里云盘链接需要 refresh_token, 请检查配置文件")
	}
	if conf.PikPak.Enable && (conf.PikPak.Username == "PIKPAK_EMAIL" || conf.PikPak.Password == "PIKPAK_PASSWORD") {
		panic("添加 PikPak 链接需要添加用户和密码, 请检查配置文件")
	}
}

// 更新 token
func UpdateToken(loginApi string, conf *models.Config) {
	loginData := models.AuthJson{
		Username: conf.Auth.Username,
		Password: conf.Auth.Password,
	}
	authJson, _ := json.Marshal(loginData)

	resData := HttpPost(loginApi, "", authJson)
	data, _ := json.Marshal(resData.Data)
	var tokenData models.AuthData
	json.Unmarshal(data, &tokenData)
	// log.Println(tokenData.Token)
	if resData.Code == 200 {
		ModConfig("config.yaml", conf, tokenData.Token)
		log.Println("token 已更新, 请重新运行此程序")
	} else {
		panic("token 更新失败, 请检查用户密码是否正确")
	}
}
