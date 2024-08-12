# 密码设置

为了安全需要，在github action设置repository security，mysqlpwd,mopwd分别设置mysql、MatrixOne的密码

# Mysql&MatrixOne设置

编辑conf下的matrixone.ini，分别设置mysql、matrixone数据库的相关信息

```
[mysql]
host='192.168.110.58'
port=3306
username='root'
# The database where test tables located
database='test'

[matrixone]
host='freetier-01.cn-hangzhou.cluster.matrixonecloud.cn'
port=6001
username='1af94fff_9a6c_4008_a946_4d15036618e1:admin:accountadmin'
# The database where test tables located
database='test'
```

# 搭建flink的k8s运行环境

使用附件创建所需的docker镜像，并推到仓库

```
docker build -t flink:1.15.4 .
```

注意修改jobmanager-session-deployment-non-ha.yaml、taskmanager-session-deployment.yaml中的镜像信息

```
spec:
  template:
    spec:
      containers:
      - name: jobmanager
        image: flink:1.15.4
```



创建k8s集群

```
#创建configmap
kubectl apply -f flink-configuration-configmap.yaml
#创建服务
kubectl apply -f jobmanager-rest-service.yaml
#部署jobmanager
kubectl apply -f jobmanager-session-deployment-non-ha.yaml
#部署taskmanager
kubectl apply -f taskmanager-session-deployment.yaml
```



执行flinkcdc任务

jobmanager服务部署采用的是nodeport，可以在浏览器中输入http://任意k8s节点:30081，访问webui，提交附件中的flink-cdc-demo-1.0.0-SNAPSHOT-jar-with-dependencies.jar任务文件，参数如下，source为Mysql的相关配置，sink为MatrixOne的相关配置：

```
--sourceHost 192.168.110.58 --sourcePort 3306 --sourceUsername root  --sourcePassword xxxxxx  --sinkHost freetier-01.cn-hangzhou.cluster.matrixonecloud.cn --sinkPort 6001 --sinkUsername 1af94fff_9a6c_4008_a946_4d15036618e1:admin:accountadmin --sinkPassword xxxxxx
```

# 建表

在Mysql、MatrixOne中分别执行如下SQL语句

```
create database test;
use test;
 CREATE TABLE `mysql_dx` (
  `id` bigint NOT NULL AUTO_INCREMENT,
  `name` varchar(100) DEFAULT NULL,
  `salary` decimal(10,0) DEFAULT NULL,
  `age` int DEFAULT NULL,
  `entrytime` date DEFAULT NULL,
  `gender` char(1) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ;
```
在mysql中建表
```
 CREATE TABLE `modify_record` (
  `id` bigint NOT NULL,
  `name` varchar(100) DEFAULT NULL,
  PRIMARY KEY (`id`)
);
```


# 创建数据&比较迁移结果

使用action定时任务触发（gen data && load data for mysql）action，完成生成数据，并把数据实时写入到mysql中，迁移完成后比较迁移结果，执行时间为utc的每天16点。当mysql和matrixone中的数据量不一致时，查看action执行日志，会有类似如下提示：

```
[Inconsistent data error]:mysql_dx table mysql num 6001215,mo num 5370379
```





 
