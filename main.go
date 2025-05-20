package main

import (
	"log"
	"ticktok-service/config"
	"ticktok-service/internal/handler"
	"ticktok-service/internal/model"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	//TIP <p>Press <shortcut actionId="ShowIntentionActions"/> when your caret is at the underlined text
	// to see how GoLand suggests fixing the warning.</p><p>Alternatively, if available, click the lightbulb to view possible fixes.</p>

	// 加载配置
	if err := config.LoadConfig("config/config.yaml"); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 初始化数据库
	if err := model.SetupDB(); err != nil {
		log.Fatalf("初始化数据库失败: %v", err)
	}

	// 设置路由
	r := handler.SetupRouter()

	// 启动服务器
	port := config.AppConfig.Server.Port
	log.Printf("服务监听端口: %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}