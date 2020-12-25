# ncmdump.go - 导出网易云音乐 NCM 格式

## 简介

用于导出网易云音乐 NCM 格式的相关内容，核心转换功能参考 [anonymous5l/ncmdump](https://github.com/anonymous5l/ncmdump)，并使用 golang 实现，起初是为了能在 Windows 下快速编译和运行。有任何BUG在[这里](https://github.com/yoki123/ncmdump/issues)提交。

## 特性
- 转换ncm文件
- 为音频(flac和mp3)文件补充tag信息，包含标题、歌手、专辑、封面等
- 保留163key使播放器能识别转换后的文件


## 如何使用？

* 下载程序[ncmdump](https://github.com/yoki123/ncmdump/releases)

 
  1. 拖拽方式执行：
   
   **拖拽ncm文件或者包含ncm文件夹到执行程序** `ncmdump-xxx`上，等待程序运行完成
   
  2. 命令行方式执行：
  
  `ncmdump-xxx [files.../dirs...]`
  参数支持：
  ```
  --output 输出文件夹，为空时默认输出文件夹为音频文件的原文件夹
  --tag    是否使用ncm的元信息来为音频文件补充tag，默认true
  ```
  参数需要放到输入文件、文件夹之前，如
  `ncmdump-xxx --output=D:\music_dump\ D:\music D:\music\name.ncm`
  


* 代码中使用

  下载：
  
```shell
  go get -u github.com/yoki123/ncmdump
```

 导入：
```golang
  import "github.com/yoki123/ncmdump"
```

顺便提一句，为了转换以及处理方便，使用 `ncmdump.Dump(fp)` 会将已经解出来的原音乐格式放入内存中，如果想直接写入文件建议修改 writer 的指向即可。

## 格式分析

NCM 实际上不是音频格式是容器格式，封装了对应格式的 Meta 以及封面等信息，主要的格式如下：

![ncm.png](./asserts/ncm.png)

因此，需要解开原格式信息的关键就是拿到 AES 的 KEY，好在每个 NCM 的加密的 KEY 都是一致的（出于性能考虑？）。所以，我们只要拿到 AES 的 KEY 以后，就可以根据格式解开对应的资源。


## 已知问题

新版的云音乐已经不在 NCM 嵌入图片以及 Meta 等信息，因此使用 `ncmdump.DumpMeta` 去调用的时候，需要检查 Meta 信息的完整性。如果您需要 Meta 等信息，建议不要使用最新的客户端。

## 相关链接

- http://www.bewindoweb.com/228.html
- https://github.com/anonymous5l/ncmdump
- https://github.com/go-flac/go-flac
- https://github.com/mingcheng/ncmdump

`- eof -`
