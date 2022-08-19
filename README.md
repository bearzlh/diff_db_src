## 数据库对比脚本

## 功能说明

### 1.对mysql.json和mysql-compare.json中配置的数据库连接进行数据对比。对比表，表字段，表数据，最后N条数据

>$ ./diff mysql -t max --parallel=100
```                                                                    
[====================================================================] 100%    0s 连接,任务(   1/   1, 361)
[====================================================================] 100%   32s 任务,异常( 361/ 361,   0)

错误统计如下
对比数据量不一致:db-1 tb1 |16900029 - 16900403| > 5
对比数据量不一致:db-1 tb2 |5520461 - 5520693| > 5
...
...
...
库名:db1     错误数:14
```

>$ ./diff mysql -h
```

可对比数据表，数据字段，数据量。可以选择某个数据库，某个数据表进行对比

可对比数据表，数据字段，数据量。可以选择某个数据表进行对比

Usage:
  diff mysql [flags]

Flags:
  -d, --dbs string      数据库连接名，多个使用英文逗号隔开，默认使用mysql.json与mysql-compare.json的同名配置对比。如果使用db1:db2的单库对比方式，db1与db2将使用mysql.json中的连接名
  -v, --debug           是否debug，默认不开启。debug会输出详细信息，包括SQL日志
  -h, --help            help for mysql
  -p, --parallel int    协程并发次数。填写过大会造成客户端&数据库压力升高 (default 4)
  -b, --tables string   可选多个，英文逗号隔开。默认全部
  -t, --types string    可选多个，英文逗号隔开。可选:table,field,max,data。分别对比表，表字段，表数据，最后N条数据 (default "table,field,max,data")
```


### 2.配置
```
{
    "debug": false, //参数中如果携带debug则会覆盖
    "max_threshold": 5, //当进行数据非停机同步时，查询延时会导致数据量不完全一致，可以设置差值在此范围属于正常情况
    "data_threshold": 5 //校验最后data_threshold条数据一致性，需要数据表包含主键
}
```

### 3.打包后的项目地址
> https://github.com/bearzlh/diff_db