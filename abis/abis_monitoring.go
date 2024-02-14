package abis

import (
	"fmt"
	"path/filepath"

	"github.com/fsnotify/fsnotify"

	"github.com/the-medium/tx-decoder/cfg"
)

func WatchDir() {

	// 파일 시스템 이벤트 모니터링 시작
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer watcher.Close()

	// 폴더 감시
	folderPath := filepath.Join(cfg.ProjectFolder, abiFilePath)
	err = watcher.Add(folderPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// 파일 추가 시 MongoDB에 저장
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				fmt.Println("new contract abi file was watched!!!")

				// MongoDB에 저장
				err := StoreAbiToABIContainer(event.Name)
				if err != nil {
					fmt.Println("Error:", err)
				}
				fmt.Println("new contract abi was registered!!!")
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Println("Error:", err)
		}
	}
}
