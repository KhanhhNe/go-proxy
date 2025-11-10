import logging
import requests

from requests.adapters import HTTPAdapter, Retry

requests.packages.urllib3.util.connection.HAS_IPV6 = False

s = requests.Session()
retries = Retry(total=1,
                backoff_factor=0.1,
                status_forcelist=[ 500, 502, 503, 504 ])

s.mount('http://', HTTPAdapter(max_retries=retries))


# logging.basicConfig()
# logging.getLogger().setLevel(logging.DEBUG)
# requests_log = logging.getLogger("requests.packages.urllib3")
# requests_log.setLevel(logging.DEBUG)
# requests_log.propagate = True

ps = {
    'ssh': 8000,
    'http': 8001,
    'socks': 8002
}

for t, port in ps.items():
    print(f"Testing {t} - {port}")

    print("HTTP - HTTP")
    print(s.get('http://api.ipify.org', proxies=dict(http=f'http://khanh:khanh@localhost:{port}')))
    print()
    s.close()

    print("HTTPS - HTTP")
    print(s.get('https://api.ipify.org', proxies=dict(https=f'http://khanh:khanh@localhost:{port}')))
    print()
    s.close()
    
    print("HTTPS - SOCKS5")
    print(s.get('https://api.ipify.org', proxies=dict(https=f'socks5://khanh:khanh@localhost:{port}')))
    print()
    s.close()

    print("HTTP - SOCKS5")
    print(s.get('http://api.ipify.org', proxies=dict(http=f'socks5://khanh:khanh@localhost:{port}')))
    print()
    s.close()