# Coffer
开发及使用说明
## 开发环境：
* Golang版本：1.16.2 linux/amd64
* Kernel版本：Linux 5.4.0-73-generic
* Ubuntu版本：Ubuntu 20.04.2 LTS
## 开发方式：
* VMware虚拟机运行Ubuntu-Server 20.04
* VS Code在Win10环境下通过SSH连接虚拟机进行远程调试
## 使用说明：
* 使用前请先将coffer可执行文件路径添加到环境变量，以下说明默认已执行该操作。

* 查看使用说明（应用中为英文版）：coffer -h或coffer -help  
* 查看版本信息：coffer -v或coffer -version
* 运行容器：coffer run <镜像名> <命令>   
其中镜像名和命令缺一不可，镜像名为可执行文件所在文件夹名称，命令为可执行文件，命令后可添加参数，如coffer run busybox sh    
  运行容器包含众多参数：   
  1. 指定容器网络：-net <网络名>
  2. 挂载数据卷：-p <宿主机目录>:<容器目录>
  3. 指定环境变量：-e <环境变量>    
  可多次调用该参数赖指定复数的环境变量。
  4. 指定容器名称：-name <容器名>
  5. 容器后台运行：-b    
  后台运行的容器将不会有任何输出到命令行界面，通过log命令查看容器日志
  6. 限制容器内存使用：-memory <限制数值>
  7. 指定网络端口映射：-port <宿主机端口>:<容器端口>
* 进入后台运行容器：coffer exec <容器名> <镜像名> <命令>    
该后台运行容器必须运行长时间运行程序
* 查看容器运行状态：coffer ps
* 查看容器日志：coffer log <容器名>
* 封装容器镜像：coffer commit <容器名> <镜像名>    
封装运行中的容器为镜像
* 停止指定容器：
coffer stop <容器名>
* 删除指定容器：
coffer rm <容器名>
* 容器网络管理：coffer network <子命令>    
  容器管理有若干子命令：     
  1. 创建网络：coffer network create -driver <驱动名> -subnet <子网地址> <网络名>    
  可指定网络驱动（目前只有网桥）和子网地址（默认 192.168.0.0/24）
  2. 显示网络列表：coffer network list
  3. 删除网络：coffer network remove <网络名>

