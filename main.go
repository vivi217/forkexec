package main

import (
	"fmt"
	"github.com/docker/docker/pkg/reexec"
	"log"
	"os"
	"strconv"
	"syscall"
	"testfork/message"
	"time"
)

var NUM = 0

const PROCESSORS = 4

func init() {
	log.Printf("init start, os.Args = %+v\n", os.Args)
	reexec.Register("childProcess", childProcess)
	if reexec.Init() {
		os.Exit(0)
	}
}

func childProcess() {
	log.Printf("childProcess[%s], os.Args = %+v\n", os.Args[1], os.Args)
	//解析参数，拿到跟父进程通信的fd
	fd, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Printf("atoi error, %s", err)
		return
	}
	for {
		time.Sleep(2 * time.Second)
		NUM = NUM + 1
		log.Printf("child process[%s], data num=[%d]\n", os.Args[1], NUM)
		var msg message.Msg
		//接收父进程发的消息
		err := message.ReadMsg(fd, &msg)
		if err != nil {
			log.Printf("read msg error. %s", err)
			return
		}
	}
}

func StartWorker() []int{
	//var fdr []int
	var fdw []int
	for i := 0; i < PROCESSORS; i++ {
		//创建父子进程通信的fd, fd是一对，分别用来收发消息
		fds, err := syscall.Socketpair(syscall.AF_LOCAL, syscall.SOCK_STREAM, 0)
		if err != nil {
			log.Panicf("Socketpair: %v", err)
		}
		//创建子进程，讲通信用的fd，通过启动参数传人子进程
		cmd := reexec.Command("childProcess", fmt.Sprintf("%d", i), fmt.Sprintf("%d", fds[0]),
			fmt.Sprintf("%d", fds[1]))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			err := CloseFD(fdw)
			if err != nil {
				log.Printf("close fdw error, %s", err)
			}
			log.Panicf("failed to run command: %s", err)
		}
		log.Println(cmd.Process.Pid)
		//fdRead := append(fdr, fds[0])
		fdw = append(fdw, fds[0])
		/*if err := cmd.Wait(); err != nil {
			log.Panicf("failed to wait command: %s", err)
		}*/
	}
	return fdw
}
func SendToWorker(fdw []int) error {
	for {
		time.Sleep(2 * time.Second)
		log.Printf("main process[%d]\n", NUM)
		msg := message.Msg{
			Title:   "socketpair msg",
			Content: "hello child",
			Size:    2,
		}
		log.Printf("main process send msg")
		for i := 0; i < PROCESSORS; i++ {
			err := message.SendMsg(fdw[i], msg)
			if err != nil {
				log.Printf("send msg error, %s", err)
				return err
			}
		}
	}
}

func CloseFD(fdw []int) error {
	for _, fd:= range fdw {
		err := syscall.Close(fd)
		if err != nil {
			log.Printf("close fd error,%s", err)
		}
	}
	return nil
}

func main() {
	log.Printf("main start, os.Args = %+v\n", os.Args)
	//启动子进程,拿到与子进程通信的fd集合
	fdw := StartWorker()
	//通过fd集合，给每个子进程发消息
	err := SendToWorker(fdw)
	if err != nil {
		log.Printf("main send msg error")
	}
	err = CloseFD(fdw)
	if err != nil {
		log.Printf("close fds error, %s", err)
	}
	log.Println("main exit")
}
