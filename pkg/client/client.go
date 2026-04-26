package client

import (
	"bufio"
	"fmt"
	"os"
	"wolf/pkg"
)

func Init() {
	var scanner = bufio.NewScanner(os.Stdin)
	fmt.Println("欢迎来到狼人杀")
	fmt.Println("=========操作菜单如下=========")
	fmt.Println("1.创建房间")
	fmt.Println("2.加入房间")
	fmt.Println("3.退出")
	fmt.Println(">请输入序号")
	index, _ := fmt.Scanln()
	switch index {
	case 1:
		player := new(pkg.Player)
		fmt.Println("\033[2J\033[H")
		fmt.Println("请输入昵称")
		name := scanner.Text()
		player.Name = name
		fmt.Println("\033[2J\033[H")
	}
}
