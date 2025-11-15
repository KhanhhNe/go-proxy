import requests

print("Downloading latest IP DB files")

for fname in ['ip-to-country.mmdb']:
    print(f"Downloading {fname}")
    folder = fname.split('.')[0]
    res = requests.get(f'https://github.com/iplocate/ip-address-databases/raw/refs/heads/main/{folder}/{fname}?download=', stream=True)
    open(f'binary/{fname}', 'wb').write(res.content)
    print(f"Downloaded {fname}")