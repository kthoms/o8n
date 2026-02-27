package main

import (
"fmt"
"log"

"github.com/kthoms/o8n/internal/app"
"github.com/kthoms/o8n/internal/config"
)

func main(){
envCfg, err := config.LoadEnvConfig("o8n-env.yaml")
if err!=nil{ log.Fatalf("load env: %v", err)}
appCfg, err := config.LoadAppConfig("o8n-cfg.yaml")
if err!=nil{ log.Fatalf("load app: %v", err)}
state, err := config.LoadAppState("o8n-stat.yml")
if err!=nil{ log.Fatalf("load state: %v", err)}
skinName := state.Skin
fmt.Printf("state.Skin=%q\n", skinName)
m := app.NewModelEnvApp(envCfg, appCfg, skinName)
fmt.Printf("m.activeSkin=%q\n", m.ActiveSkin())
if m.Skin()==nil{
fmt.Println("m.skin == nil")
} else {
fmt.Printf("m.skin.Color(\"borderFocus\")=%q\n", m.Skin().Color("borderFocus"))
fmt.Printf("m.skin.Color(\"fg\")=%q\n", m.Skin().Color("fg"))
}
}
