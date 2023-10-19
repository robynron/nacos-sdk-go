package encryption

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/common/logger"
	"strings"
	"sync"
)

var (
	initDefaultHandlerOnce = &sync.Once{}
	defaultHandler         *DefaultHandler
)

type HandlerParam struct {
	DataId           string `json:"dataId"`  //required
	Content          string `json:"content"` //required
	EncryptedDataKey string `json:"encryptedDataKey"`
	PlainDataKey     string `json:"plainDataKey"`
	KeyId            string `json:"keyId"`
}

type Plugin interface {
	Encrypt(*HandlerParam) error
	Decrypt(*HandlerParam) error
	AlgorithmName() string
	GenerateSecretKey(*HandlerParam) (string, error)
	EncryptSecretKey(*HandlerParam) (string, error)
	DecryptSecretKey(*HandlerParam) (string, error)
}

type Handler interface {
	EncryptionHandler(*HandlerParam) error
	DecryptionHandler(*HandlerParam) error
	RegisterPlugin(Plugin) error
}

func GetDefaultHandler() Handler {
	if defaultHandler == nil {
		initDefaultHandler()
	}
	return defaultHandler
}

func initDefaultHandler() {
	initDefaultHandlerOnce.Do(func() {
		defaultHandler = &DefaultHandler{
			encryptionPlugins: make(map[string]Plugin, 2),
		}
	})
}

type DefaultHandler struct {
	encryptionPlugins map[string]Plugin
}

func (d *DefaultHandler) EncryptionHandler(param *HandlerParam) error {
	if err := d.encryptionParamCheck(*param); err != nil {
		if err == DataIdParamCheckError || err == ContentParamCheckError {
			return nil
		}
		return err
	}
	plugin, err := d.getPluginByDataIdPrefix(param.DataId)
	if err != nil {
		return err
	}
	plainSecretKey, err := plugin.GenerateSecretKey(param)
	if err != nil {
		return err
	}
	param.PlainDataKey = plainSecretKey
	return plugin.Encrypt(param)
}

func (d *DefaultHandler) DecryptionHandler(param *HandlerParam) error {
	if err := d.decryptionParamCheck(*param); err != nil {
		if err == DataIdParamCheckError || err == ContentParamCheckError {
			return nil
		}
		return err
	}
	plugin, err := d.getPluginByDataIdPrefix(param.DataId)
	if err != nil {
		return err
	}
	plainSecretkey, err := plugin.DecryptSecretKey(param)
	if err != nil {
		return err
	}
	param.PlainDataKey = plainSecretkey
	return plugin.Decrypt(param)
}

func (d *DefaultHandler) getPluginByDataIdPrefix(dataId string) (Plugin, error) {
	var (
		matchedCount  int
		matchedPlugin Plugin
	)
	for k, v := range d.encryptionPlugins {
		if strings.Contains(dataId, k) {
			if len(k) > matchedCount {
				matchedCount = len(k)
				matchedPlugin = v
			}
		}
	}
	if matchedPlugin == nil {
		return matchedPlugin, PluginNotFoundError
	}
	return matchedPlugin, nil
}

func (d *DefaultHandler) RegisterPlugin(plugin Plugin) error {
	if _, v := d.encryptionPlugins[plugin.AlgorithmName()]; v {
		logger.Warnf("encryption algorithm [%s] has already registered to defaultHandler, will be update", plugin.AlgorithmName())
	} else {
		logger.Infof("register encryption algorithm [%s] to defaultHandler", plugin.AlgorithmName())
	}
	d.encryptionPlugins[plugin.AlgorithmName()] = plugin
	return nil
}

func (d *DefaultHandler) encryptionParamCheck(param HandlerParam) error {
	if err := d.dataIdParamCheck(param.DataId); err != nil {
		return DataIdParamCheckError
	}
	if err := d.contentParamCheck(param.Content); err != nil {
		return ContentParamCheckError
	}
	if err := d.keyIdParamCheck(param.Content); err != nil {
		return KeyIdParamCheckError
	}
	return nil
}

func (d *DefaultHandler) decryptionParamCheck(param HandlerParam) error {
	if err := d.dataIdParamCheck(param.DataId); err != nil {
		return DataIdParamCheckError
	}
	if err := d.contentParamCheck(param.Content); err != nil {
		return ContentParamCheckError
	}
	return nil
}

func (d *DefaultHandler) dataIdParamCheck(dataId string) error {
	if !strings.Contains(dataId, CipherPrefix) {
		return fmt.Errorf("dataId prefix should start with: %s", CipherPrefix)
	}
	return nil
}

func (d *DefaultHandler) contentParamCheck(content string) error {
	if len(content) == 0 {
		return fmt.Errorf("content need to encrypt is nil")
	}
	return nil
}

func (d *DefaultHandler) keyIdParamCheck(keyId string) error {
	if len(keyId) == 0 {
		return fmt.Errorf("keyId is nil, need to be set")
	}
	return nil
}
