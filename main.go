package main

import (
    // "time"
	"context"
	"fmt"
	"log"
    "bytes"
    "net/http"
    "io/ioutil"
    "encoding/json"

	"github.com/ericchiang/k8s"
)
type Datasource struct {
    Name string `json:"name"`
    Type string `json:"type"`
    Access string `json:"access"`
    Url string `json:"url"`
}

type Datasources struct {
    Items []Datasource `json:"datasources"`
}

func (ds *Datasources) AddDatasource(datasource Datasource) []Datasource {
    ds.Items = append(ds.Items, datasource)
    return ds.Items
}

func buildList() Datasources{
	client, err := k8s.NewInClusterClient()
    var dss Datasources
	if err != nil {
		log.Fatal(err)
	}
	ls := new(k8s.LabelSelector)
	ls.Eq("app", "prometheus")

	pods, err := client.CoreV1().ListPods(context.Background(), k8s.AllNamespaces, ls.Selector())
	if err != nil {
		log.Fatal(err)
	}
    fmt.Println("damn")
	for _, pod := range pods.Items {
        dns := *pod.Metadata.Name + "." + *pod.Metadata.Namespace + ".svc.cluster.local"
        ds := Datasource{Name: dns, Type: "prometheus", Access: "proxy", Url: dns}
        dss.AddDatasource(ds)
        // fmt.Println("+%v", dns)
        // fmt.Println("WTF")
        // fmt.Println("+%v", ds)
	}
    // for _, d := range dss.Items {
    //     fmt.Println(d)
    // }

    return dss
}

func main() {
    fmt.Println("What")
    dss := buildList()
    for _, d := range dss.Items {
        fmt.Println(d)
    }
    // time.Sleep(60 * time.Second)
    url := "http://localhost:3000/api/datasources"
    b, err := json.Marshal(dss)
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

    fmt.Printf("%+v", dss)
}
