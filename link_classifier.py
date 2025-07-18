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


def test_main():
    """测试主函数的功能"""
    # 测试链接
    test_links = [
        "https://console.cloud.baidu-int.com/devops/icafe/issue/OnlineApi-21205/show?source=copy-shortcut",
        "https://ku.baidu-int.com/knowledge/HFVrC7hq1Q/iki-aVkbF_/9sd-6XkDV8/zRCXYKbDS6IZGF"
    ]
    
    print("链接分类测试:")
    print("=" * 50)
    
    for link in test_links:
        result = main(url=link)  # 明确指定参数名
        print(f"链接: {link}")
        print(f"类型: {result['result']}")
        print("-" * 50)
    
    # 交互式输入测试
    print("\n请输入链接进行测试 (输入'quit'退出):")
    while True:
        user_input = input("链接地址: ").strip()
        if user_input.lower() == 'quit':
            break
        if user_input:
            result = main(url=user_input)  # 明确指定参数名
            print(f"结果: {result['result']}\n")


if __name__ == "__main__":
    test_main() 