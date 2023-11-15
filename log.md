# Day 1:
先了解了一下Go语言的相关背景知识，发现它跟C语言渊源还挺深，给人第一感觉就像是C语言和python语言结合产生的。  
先配了个Go语言的开发环境，使用vscode，安装go的相关插件，添加了环境变量，插件很贴心地下载了很多工具  
目前开发环境拥有基本的代码高亮，自动补充与调试功能。  
写了个测试程序，运行结果如下：
![Alt text](image.png)  
期间也遇上了不少问题，但其中最突出的一个还是刚写完程序提示错误  
    
    gopls was not able to find modules in your workspace.

    When outside of GOPATH, gopls needs to know which modules you are working on.

    You can fix this by opening your workspace to a folder inside a Go module, or

    by using a go.work file to specify multiple modules.

    See the documentation for more information on setting up your workspace
还挺让人不解的。  
之后经过搜索发现这涉及一个golang 1.11 新加的go modules特性，要在工作区中设置一个go.mod文件来指定工作区范围。编程小白世面见得确实少。根据资料使用go mod 命令解决问题。  
![Alt text](image-1.png)  

# Day 2:
上一天课。  
真该死啊，计网实验课  

# Day 3:
简单看了一下Go的语法，光看还是没什么用    
干脆直接上手  
先申请得到AK，塞进环境变量  
查阅资料得知 Go里通过os.Getenv()获取环境变量  
有一个完善的 net/http 包，通过 net/http 包可以很方便的搭建一个可以运行的 Web 服务器。  
参考官网给出的示例代码,先打出一个框架  
其他部分的语法和库的使用还得再熟悉一下,json的知识还得补一补

# Day 4:
简单实现了对返回数据的解析  
进行测试
![Alt text](image-2.png)
结果
![Alt text](image-3.png)

# Day 5:
完成了对post参数的处理
![Alt text](image-4.png)
虽然不知道效果是否达到预期  
着手对Task 2进行实现

# Day 6:
完成了对Task 2的实现,初步书写了地理编码，但是在反序列化的时候出了点问题  
发现原来是因为返回的数据并不是纯json，通过修改请求参数解决  
挺让我疑惑的点在于对于步行等交通方式真的有讨论路况的必要吗

# Day 7:
简单了解了database/sql库,但这个“名字A->全名B”的映射让人难以实现，不清楚是要使用户输入一个名字后由程序自动转换为常用名字，还是让用户在输入的过程中通过表单下拉框自动进行提示。  
直觉上讲应该是要利用下拉框自动进行提示但这样前端难度反而更大  
总之先走一步算一步把数据库API调用写出来，实现一个简单的历史记录查询

# Day 8:
了解了ORM的基本思想，大体上看是把数据库中的记录以对象的形式来记录，进行增删改查时直接调用方法而不是使用数据库的API，
表 -> 类  
行 -> 对象  
列 -> 属性 
因此先简单学习了结构体的一些用法  
发现go语言中不使用对象，而是用结构体来代替类，在声明时设置接收者来实现由特定结构体类型的变量使用的函数，来代替this  
注意到如果函数中的接收者仅仅只是作为一个值传递的参数，对它的属性进行修改是无法影响到实际变量的，  
查询得知此时接收者应该为结构指针，而同时在使用时可以省略"*",这倒是第一次见  

# Day 9:
一天都没什么时间  
因为用对象来操作数据库涉及到很多的方法调用，但注意到如果由函数来声明一个db连接池再返回指针，那部分局部变量就应该要被释放了  
如果不把开启数据库操作封装起来又显得很臃肿，之前在写C语言时还可以用malloc申请个堆空间，结果go里没有，  
一查发现go很贴心的识别出了逃逸的变量，把它自动放入了堆  

# Day 10:
眼瞅着最后一天就到了，几个task写得缺胳膊少腿，但往好处想起码跑得起来  
本来数据库的host和password也用环境变量来获取  
不知出了什么毛病居然得到都是空字符串  
表名为history  
两个字段 origin destination 均为字符串  
剩下个删除操作方法写好了，但还没用上。
