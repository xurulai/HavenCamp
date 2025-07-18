const { defineConfig } = require('@vue/cli-service')
const fs = require('fs')
const path = require('path')

// 检查是否在Docker环境中构建
const isDocker = process.env.DOCKER_BUILD === 'true'

// 创建devServer配置
let devServerConfig = {
  host: '0.0.0.0',
  port: 8080,
}

// 只有在非Docker环境且SSL证书存在时才启用HTTPS
if (!isDocker && fs.existsSync("/etc/ssl/certs/server.crt") && fs.existsSync("/etc/ssl/private/server.key")) {
  devServerConfig = {
    host: '0.0.0.0',
    https: {
      cert: fs.readFileSync("/etc/ssl/certs/server.crt"),
      key: fs.readFileSync("/etc/ssl/private/server.key"),
    },
    port: 443,
  }
}

module.exports = defineConfig({
  transpileDependencies: true,
  lintOnSave: false,
  devServer: devServerConfig
})
