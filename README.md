# UPDATE

Starting from [version 1.26](https://github.com/serverless/serverless/releases/tag/v1.26.0) Serverless Framework includes two Golang templates:

* `aws-go` - basic template with two functions
* `aws-go-dep` - **recommended** template using [`dep`](https://github.com/golang/dep) package manager

You can use them with `create` command:

```
serverless create -t aws-go-dep
```

Original README below.

---

# Serverless Template for Golang

This repository contains template for creating serverless services written in Golang.

## Quick Start

1. Create a new service based on this template

```
serverless create -u https://github.com/serverless/serverless-golang/ -p myservice
```

2. Compile function

```
cd myservice
GOOS=linux go build -o bin/main
```

3. Deploy!

```
serverless deploy
```

4. ポリシーをアタッチ
- このままだと` is not authorized to perform: ce:GetCostAndUsage on resource: arn:aws:ce:us-east-1:`エラーが出るのでポリシーをアタッチする
### 4-1：ポリシーの作成
- IAM→ポリシーより`ポリシーの作成`を選択し以下の内容で作成する
  - サービス：Cost Explorer Service
  - アクション：GetCostAndUsage
  - ポリシーの名前：任意（AmazonCostExplorerGetCostAccessとした）
  
### 4-2：lambdaに割り当てたロールに4-1で作成したポリシーをアタッチする
- ロール（基本プロジェクト名 + us-east-1みたいな感じで自動生成される）を選択して4-1で作成したポリシーをアタッチする

