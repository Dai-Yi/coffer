# Coffer
### 毕业设计 相关笔记
## 隔离Namespace相关： 
### CLONE_NEWUTS (UTS Namespace)：隔离hostname 和NIS domain name。  
hostname：主机名，在局域网中标记主机。  
NIS domain name：NIS域名，在NIS服务器中标记主机。  
* 验证方法：  
    1. 修改hostname：hostname -b xxx  
    2. 在新终端中查看hostname  
### CLONE_NEWPID (PID Namespace)：隔离进程ID。  
* 验证方法：  
    1. 查看进程树：pstree -pl  
    2. 在新终端中查看进程真实PID：pstree -pl  
### CLONE_NEWIPC (IPC Namespace)：隔离进程间通信，System V IPC和POSIX message queues。  
System V IPC 、POSIX message queues：IPC包含共享内存、信号量和消息队列，用于进程间通信。每个IPC名称空间都有其自己的一组System V IPC标识符和它自己的POSIX 消息队列文件系统。  
* 验证方法：  
    1. 查看现有的IPC消息序列：ipcs -q  
    2. 创建一个IPC消息序列：ipcmk -Q  
    3. 再查看当前IPC消息序列：ipcs -q  
    4. 在新终端中查看IPC消息序列：ipcs -q  
### CLONE_NEWNS (Mount Namespace)：隔离进程挂载点。  
* 验证方法：  
    1. 查看/proc：ls /proc  
    2. 在新终端中查看/porc：ls /proc  
    3. 查看系统进程：ps -ef  
    4. 在新终端中查看系统进程：ps -ef  
### CLONE_NEWNET (Network Namespace)：隔离网络设备。  
* 验证方法：  
    1. 查看网络设备情况：ifconfig  
    2. 在新终端查看网络设备状况：ifconfig  
## 限制资源Cgroups相关：  
### 限制内存资源（待更新）  
* 验证方法：  
    1. 使用stress程序占用200m内存：stress --vm-bytes 200m --vm-keep -m 1  
    2. 在新终端中查看当前系统资源占用情况：top  
    3. 关闭top终端  
    4. 用容器中打开stress程序，同时容器限制100m内存：coffer run -s -memory 100m stress --vm-bytes 200m --vm-keep -m -1  
    5. 在新终端中查看此时系统资源占用情况：top  
## 数据卷相关：持久化容器数据  
* 验证方法：  
    1. 查看root目录下文件：ls /root  
    2. 启动容器，使宿主机某目录作为数据卷使容器数据持久化：coffer run -s -d /root/volume:/containerVolume sh  
    3. 查看容器目录下文件：ls   
    4. containerVolume目录下创建HelloWorld文件：touch HelloWorld.txt  
    5. 在新终端中查看root目录下文件：ls /root  
    6. 查看volume目录下文件：ls /root/volume  
## /proc相关：
### /proc/[pid]/mountinfo：
文件，挂载信息，格式为36 35 98:0 /mnt1 /mnt2 rw,noatime master:1 - ext3 /dev/root rw,errors=continue，以空格作为分隔符，从左到右各字段的意思分别是唯一挂载ID、父挂载ID、文件系统的设备主从号码、文件系统中挂载的根节点、相对于进程根节点的挂载点、挂载权限等挂载配置、可选配置、短横线表示前面可选配置的结束、文件系统类型、文件系统特有的挂载源或者为none、额外配置。
### /proc/self：
目录，链接到了当前进程所在的目录  
### /proc/[pid]/cgroup：
文件，进程所属的控制组，格式为冒号分隔的三个字段，分别是结构ID、子系统、控制组，需配置CONFIG_CGROUPS。  
## Mount相关：
* MS_PRIVATE：挂载点设为私有。  
* MS_REC：递归更改子树中安装的传播类型。 
* MS_NOEXEC：本文件系统不允许运行其他程序。  
* MS_NOSUID：本系统运行时不允许更改用户ID和组ID。  
* MS_NODEV：不允许访问此文件系统上的设备（特殊文件）。  
* MS_BIND：绑定安装，绑定安装使文件或目录子树在单个目录层次结构的另一点可见。  
## Umount相关：
* MNT_DETACH：使挂载点不可用于新访问，立即相关文件系统彼此以及与挂载表断开连接。  

