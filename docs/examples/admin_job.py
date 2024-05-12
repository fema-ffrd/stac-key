import requests
import json
import os
from headers import get_auth_header
from urllib3.exceptions import InsecureRequestWarning
from urllib3 import disable_warnings

disable_warnings(InsecureRequestWarning)
from dotenv import load_dotenv, find_dotenv

load_dotenv(find_dotenv())

headers = get_auth_header()

http_verb = "POST"
test_url = f"{os.getenviron('API_URL')}/collections"
payload = {"id": "new-collection"}

response = requests.request(
    http_verb, test_url, headers=headers, data=json.dumps(payload), verify=False
)

print(response.text)