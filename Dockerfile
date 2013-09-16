from ubuntu:12.04
maintainer Nick Stinemates

run apt-get install -y python-setuptools
run easy_install pip

add . /website
run pip install -r /website/requirements.txt
env PYTHONPATH /website
expose 5000

cmd ["python", "website/web/server.py"]
