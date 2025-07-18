#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import json
from urllib.parse import urlparse

def main(url: str) -> dict:
    """
    解析URL链接，返回链接类型（icafe或iku）
    
    Args:
        url (str): 需要识别的URL链接
        
    Returns:
        dict: 包含result键的字典，result为链接类型字符串
    """
    try:
        # 检查输入数据
        if not url:
            return {
                'result': ""
            }
        
        # 解析URL
        parsed_url = urlparse(url)
        
        # 根据域名前缀判断链接类型
        if parsed_url.netloc == 'console.cloud.baidu-int.com':
            link_type = "icafe"
        elif parsed_url.netloc == 'ku.baidu-int.com':
            link_type = "iku"
        else:
            link_type = ""
        
        return {
            'result': link_type
        }
        
    except Exception as e:
        print(f"解析URL时出错: {str(e)}")
        return {
            'result': ""
        }

# 测试
if __name__ == "__main__":
    # 测试icafe链接
    icafe_url = "https://console.cloud.baidu-int.com/devops/icafe/issue/OnlineApi-21205/show?source=copy-shortcut"
    result1 = main(icafe_url)
    print(f"ICafe链接测试: {result1}")
    
    # 测试iku链接
    iku_url = "https://ku.baidu-int.com/knowledge/HFVrC7hq1Q/iki-aVkbF_/9sd-6XkDV8/zRCXYKbDS6IZGF"
    result2 = main(iku_url)
    print(f"IKU链接测试: {result2}")
    
    # 测试空链接
    result3 = main("")
    print(f"空链接测试: {result3}")
    
    # 测试无效链接
    result4 = main("https://www.baidu.com")
    print(f"无效链接测试: {result4}") 