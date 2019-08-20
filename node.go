package gmq

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Node struct {
	HttpAddr string `json:"http_addr"`
	TcpAddr  string `json:"tcp_addr"`
	JoinTime string `json:"join_time"`
	Weight   int    `json:"weight"`
}

func GetNodes(registerAddr string) ([]*Node, error) {
	var nodes []*Node
	ts := strings.Split(registerAddr, ",")
	for _, t := range ts {
		url := fmt.Sprintf("%s/getNodes", t)
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		v := struct {
			Code int `json:"code"`
			Data struct {
				Nodes []*Node `json:"nodes"`
			} `json:"data"`
			Msg string `json:"msg"`
		}{}

		if err := json.Unmarshal(data, &v); err != nil {
			return nil, err
		}
		nodes = append(nodes, v.Data.Nodes...)
	}

	return nodes, nil
}

func RemoveNode(nodes []*Node, index int) []*Node {
	nodesLen := len(nodes)
	if nodesLen == 0 {
		return nil
	} else if nodesLen == 1 {
		return nodes[0:0]
	} else {
		return append(nodes[0:index], nodes[index+1:]...)
	}
}
