package vk

import "sync"

type VkAppsPull struct {
	locker  sync.Locker
	callbackUrl string
	secrets map[string]string
	tokens  map[string]AccessToken
}

var vkAppsPull VkAppsPull = VkAppsPull{}


func RegisterVkApps(callbackUrl string, appSecrets map[string]string) {

}

func (pull VkAppsPull) addToken(appID string, token AccessToken) {
	pull.locker.Lock()
	pull.tokens[appID] = token
	pull.locker.Unlock()
}

//func (pull VkAppsPull) GetToken(appID, secret string) AccessToken {
//	if token, ok := pull.tokens[appID]; ok {
//		return token
//	}
//	api := NewAPI(appID, secret, nil, "")
//	api.Authenticate()
//}