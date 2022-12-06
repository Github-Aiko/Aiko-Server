package api

import (
	"testing"
)

func CreateClient() *Client {
	apiConfig := &Config{
		APIHost:  "http://127.0.0.1:8080",
		NodeID:   4,
		Token:    "123456789123456789",
		NodeType: "Trojan",
	}
	client := New(apiConfig)
	return client
}

func TestGetNodeInfo(t *testing.T) {
	client := CreateClient()
	nodeInfo, err := client.GetNodeInfo()
	if err != nil {
		t.Error(err)
	}
	t.Log(nodeInfo)
}

func TestGetUserList(t *testing.T) {
	client := CreateClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}
	t.Log(userList)
}

func TestReportUserTraffic(t *testing.T) {
	client := CreateClient()
	userList, err := client.GetUserList()
	if err != nil {
		t.Error(err)
	}
	generalUserTraffic := make([]*UserTraffic, len(*userList))
	for i, userInfo := range *userList {
		generalUserTraffic[i] = &UserTraffic{
			UID:      userInfo.ID,
			Upload:   114514,
			Download: 114514,
		}
	}
	//client.Debug()
	err = client.ReportUserTraffic(generalUserTraffic)
	if err != nil {
		t.Error(err)
	}
}
