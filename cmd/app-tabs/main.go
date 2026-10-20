package main

import (
 "context"
 "log"
 "os"
 "os/signal"
 "syscall"
 "github.com/MiracleOS-Team/desktop/internal/apptabs"
)
func main(){ctx,cancel:=signal.NotifyContext(context.Background(),os.Interrupt,syscall.SIGTERM);defer cancel();if err:=apptabs.Run(ctx);err!=nil{log.Fatal(err)}}
