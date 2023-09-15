package main

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"io/ioutil"
	"os"
	"path"
	"time"
)

var localServerConfigWithOptions = constant.NewServerConfig(
	"mse-b3c2b0a2-p.nacos-ans.mse.aliyuncs.com",
	8848,
)

var localClientConfigWithOptions = constant.NewClientConfig(
	constant.WithTimeoutMs(10*1000),
	constant.WithBeatInterval(2*1000),
	constant.WithNotLoadCacheAtStart(true),
	constant.WithAccessKey(getFileContent(path.Join(getWDR(), "ak"))),
	constant.WithSecretKey(getFileContent(path.Join(getWDR(), "sk"))),
	constant.WithNamespaceId("791fd262-3735-40df-a605-e3236f8ff495"),
	constant.WithOpenKMS(true),
	constant.WithKMSVersion(constant.KMSv3),
	constant.WithKMSv3Config(&constant.KMSv3Config{
		ClientKeyContent: getFileContent(path.Join(getWDR(), "client_key.json")),
		Password:         getFileContent(path.Join(getWDR(), "password")),
		Endpoint:         getFileContent(path.Join(getWDR(), "endpoint")),
		CaContent:        getFileContent(path.Join(getWDR(), "ca.pem")),
	}),
	constant.WithRegionId("cn-beijing"),
)

var localConfig = vo.ConfigParam{
	DataId:   "cipher-crypt",
	Group:    "default",
	Content:  "crypt",
	KmsKeyId: "key-xxxwv4", //可以识别
}

func main() {
	usingKMSv3ClientAndStoredByNacos()
}

func usingKMSv3ClientAndStoredByNacos() {
	client := createConfigClient()
	published, err := client.PublishConfig(vo.ConfigParam{
		DataId:   localConfig.DataId,
		Group:    localConfig.Group,
		Content:  localConfig.Content,
		KmsKeyId: localConfig.KmsKeyId,
	})

	if published {
		fmt.Println("successfully publish content: " + localConfig.Content)
	} else {
		fmt.Println("failed to publish content: " + localConfig.Content + ", with error: " + err.Error())
	}

	time.Sleep(1 * time.Second)

	content, err := client.GetConfig(vo.ConfigParam{
		DataId: localConfig.DataId,
		Group:  localConfig.Group,
	})
	if err != nil {
		fmt.Println("failed with err:" + err.Error())
	}
	fmt.Println("successfully get content:" + content)

}

func onlyUsingKMSv3() error {
	client := createConfigClient()
	encrypt, err := client.KMSv3Encrypt(localConfig.DataId, localConfig.Content, localConfig.KmsKeyId)
	if err != nil {
		return err
	}
	fmt.Println("encrypt : " + string(encrypt))
	decrypt, err := client.KMSv3Decrypt(localConfig.DataId, encrypt)
	if err != nil {
		return err
	}
	fmt.Println("decrypt : " + string(decrypt))
	return nil
}

func createConfigClient() *config_client.ConfigClient {
	nc := nacos_client.NacosClient{}
	_ = nc.SetServerConfig([]constant.ServerConfig{*localServerConfigWithOptions})
	_ = nc.SetClientConfig(*localClientConfigWithOptions)
	_ = nc.SetHttpAgent(&http_agent.HttpAgent{})
	client, err := config_client.NewConfigClient(&nc)
	if err != nil {
		fmt.Println("create config client failed: " + err.Error())
	}
	return client
}

func getWDR() string {
	getwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return getwd
}

func getFileContent(filePath string) string {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}
	return string(file)
}
