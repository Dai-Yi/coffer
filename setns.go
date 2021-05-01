package main

/*
#define _GNU_SOURCE
#include <unistd.h>
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
//
//该文件是cgo文件，注释中的代码不能删
//
//__attribute__是编译属性，用于向编译器描述特殊的标识、错误检查或高级优化
//__attribute__((constructor)) 说明此函数在 在main函数之前调用
__attribute__((constructor)) void enterNamespace(void) {
	char *cofferPID;
	//从环境变量中获取pid
	cofferPID = getenv("cofferPID");
	if (cofferPID) {
		//fprintf(stdout, "got cofferPID=%s\n", cofferPID);
	} else {//若没有发现指定pid则退出
		//fprintf(stdout, "missing cofferPID env skip nsenter");
		return;
	}
	char *cofferCMD;
	//从环境变量中获取需要执行的命令
	cofferCMD = getenv("cofferCMD");
	if (cofferCMD) {
		//fprintf(stdout, "got cofferCMD=%s\n", cofferCMD);
	} else {//若没有发现指定命令则退出
		//fprintf(stdout, "missing cofferCMD env skip nsenter");
		return;
	}
	int i;
	char namespacePath[1024];
	char *namespaces[] = { "ipc", "uts", "net", "pid", "mnt" };//需要进入的5种namespace
	for (i=0; i<5; i++) {
		//拼接对应路径,/proc/pid/ns/namespace
		sprintf(namespacePath, "/proc/%s/ns/%s", cofferPID, namespaces[i]);
		int fd = open(namespacePath, O_RDONLY);
		//使用setns系统调用进入对应的namespace
		if (setns(fd, 0) == -1) {
			//fprintf(stderr, "setns on %s namespace failed: %s\n", namespaces[i], strerror(errno));
		} else {
			//fprintf(stdout, "setns on %s namespace succeeded\n", namespaces[i]);
		}
		close(fd);
	}
	int res = system(cofferCMD);//在进入的namespace种执行指定的命令
	exit(0);
	return;
}
*/
import "C"
