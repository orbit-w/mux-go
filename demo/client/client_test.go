package main

import (
	"fmt"
	"github.com/spf13/viper"
	"testing"
)

/*
   @Author: orbit-w
   @File: client_test
   @2024 8月 周四 01:31
*/

func Test_config(t *testing.T) {
	parseConfig()
	fmt.Println(viper.GetString("v"))
	fmt.Println(viper.GetString("log_dir"))
}
