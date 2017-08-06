// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ericchiang/k8s"
)

type Datasource struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Access string `json:"access"`
	Url    string `json:"url"`
}

func getWatch() *k8s.CoreV1ServiceWatcher {
	client, err := k8s.NewInClusterClient()
	if err != nil {
		log.Fatal(err)
	}
	ls := new(k8s.LabelSelector)
	ls.Eq("app", "prometheus")

	servicewatch, err := client.CoreV1().WatchServices(context.Background(), k8s.AllNamespaces, ls.Selector())
	if err != nil {
		log.Fatal(err)
	}
	return servicewatch
}
func watchLoop(servicewatch *k8s.CoreV1ServiceWatcher) {
	for {
		eventtype, pod, err := servicewatch.Next()
		if err != nil {
			log.Fatal(err)
			break
		}
		podname := *pod.Metadata.Name + "." + *pod.Metadata.Namespace
		dns := podname + ".svc.cluster.local"
		if *eventtype.Type == "DELETED" {
			RemoveDataSource(podname)
		} else {
			if *eventtype.Type == "ADDED" {
				url := "http://" + dns + ":9090"
				ds := Datasource{Name: podname, Type: "prometheus", Access: "proxy", Url: url}
				AddDataSource(ds)
			}
		}
	}
}

func main() {
	servicewatch := getWatch()
	watchLoop(servicewatch)
}
func RemoveDataSource(podname string) {
	fmt.Println("--- DELETE SERVICE ---")
	url := "http://localhost:3000/api/datasources/name/" + podname
	req, _ := http.NewRequest("DELETE", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	var bufferDelete bytes.Buffer
	resp.Write(&bufferDelete)
	fmt.Println(string(bufferDelete.Bytes()))
}
func AddDataSource(ds Datasource) {
	fmt.Println("--- ADD SERVICE ---")
	url := "http://localhost:3000/api/datasources"
	b, err := json.Marshal(ds)
	if err != nil {
		fmt.Println(err)
		return
	}
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(b))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
}
