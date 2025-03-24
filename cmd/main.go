package main

import (
	"fmt"
	"os"

	"github.com/KangYoungIn/httpeek/internal/handler"
	typ "github.com/KangYoungIn/httpeek/internal/type"
	"github.com/KangYoungIn/httpeek/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

var (
	method      string
	headers     []string
	body        string
	showHeaders bool
	showBody    bool
	apiMode     bool
	apiPort     string
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "httpeek [url...]",
		Short: "HTTP 요청 흐름을 추적하고 시각화합니다",
		Run: func(cmd *cobra.Command, args []string) {
			if apiMode {
				startAPIServer()
				return
			}

			if len(args) == 0 {
				fmt.Fprintln(os.Stderr, "URL이 지정되지 않았습니다.")
				cmd.Usage()
				os.Exit(1)
			}

			for _, url := range args {
				fmt.Println("==================================================")
				fmt.Printf("분석 대상: %s\n", url)
				fmt.Println("==================================================")

				config := typ.TraceConfig{
					Method:      method,
					Headers:     headers,
					Body:        body,
					ShowHeaders: showHeaders,
					ShowBody:    showBody,
				}

				if err := utils.TraceURL(url, config); err != nil {
					fmt.Fprintf(os.Stderr, "오류 (%s): %v\n", url, err)
				}
			}
		},
	}

	rootCmd.Flags().BoolVar(&apiMode, "api", false, "API 서버 모드로 실행")
	rootCmd.Flags().StringVar(&apiPort, "port", "8080", "API 서버 포트 (기본: 8080)")

	rootCmd.Flags().StringVarP(&method, "method", "X", "GET", "HTTP 메서드 (GET, POST, etc)")
	rootCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "요청 헤더 (예: -H \"Key: Value\")")
	rootCmd.Flags().StringVar(&body, "body", "", "요청 본문 (JSON/text)")
	rootCmd.Flags().BoolVar(&showHeaders, "show-headers", false, "응답 헤더 출력")
	rootCmd.Flags().BoolVar(&showBody, "show-body", false, "응답 Body 출력")

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "실행 오류: %v\n", err)
		os.Exit(1)
	}
}

func startAPIServer() {
	fmt.Println("API 서버 시작 중...")

	r := gin.Default()
	r.POST("/trace", handler.TraceHandler)
	if err := r.Run(":" + apiPort); err != nil {
		fmt.Fprintf(os.Stderr, "API 서버 실행 실패: %v\n", err)
		os.Exit(1)
	}
}
