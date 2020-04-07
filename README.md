# grafana-chs

**汉化分非 master 请切换到汉化分支：v6.3.4-chs。**

## grafana 的前端与后端均以汉化处理。
* 前端目录为 public
* 后端目录为 pkg

## 部署
* 前端替换原程序或在配置中设置静态目录路径。
* 后端需要通过go run build.go build 生成bin文件(请注意生成的系统，windows默认生成windows程序，linux，mac等类似)。

*目前汉化工作比较粗浅，可能对具体名次，名称出现差错或汉化不完全等问题。*


## 图片展示：
* 前端：
![](https://s1.ax1x.com/2020/04/07/GceBuR.jpg)
![](https://s1.ax1x.com/2020/04/07/GceDD1.png)
![](https://s1.ax1x.com/2020/04/07/Gcea34.png)
*可以看到，菜单和页面标题是英文的，这需要在后端进行处理。*

* 后端：
![](https://s1.ax1x.com/2020/04/07/GcerHx.png)
![](https://s1.ax1x.com/2020/04/07/Gcewv9.jpg)
以警报为例，可以看到后端汉化后，页面才是完整的中文化。

* 个别插件汉化：
![](https://s1.ax1x.com/2020/04/07/GcmzSe.jpg)
插件汉化比较繁琐，目前汉化不是很完全。尤其是单位选择那些（太TM多了 —。—｜｜｜）。
