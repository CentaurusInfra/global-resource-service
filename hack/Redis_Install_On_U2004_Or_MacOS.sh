#!/usr/bin/bash
#
# This script is used to quickly install configure Redis Server 
# on Ubuntun20.04 and MacOS(Darwin 20.6.0)
#    Running on Ubuntu: ./hack/Redis_On_U2004_Or_MacOs.sh
#
#    Running on MacOS:  /bin/bash ./hack/Redis_On_U2004_Or_MacOs.sh
#
# Reference: 
#    For Ubuntu: https://redis.io/docs/getting-started/installation/install-redis-on-linux/
#                https://www.linode.com/docs/guides/install-redis-ubuntu/
#
#    For MacOS:  https://redis.io/docs/getting-started/installation/install-redis-on-mac-os/
#    
#
export PATH=$PATH

if [ `uname -s` == "Linux" ]; then
  LinuxOS=`uname -v |awk -F'-' '{print $2}' |awk '{print $1}'`

  if [ "$LinuxOS" == "Ubuntu" ]; then
    echo "1. Install Redis on Ubuntu ......"
    curl -fsSL https://packages.redis.io/gpg | sudo gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg

    echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/redis.list

    sudo apt-get update
    sudo apt-get install redis
    
    echo "End to install on Ubuntu ......"

    echo ""
    echo "2. Enable and Run Redis ......"
    echo "==============================="
    REDIS_CONF_Ubuntu=/etc/redis/redis.conf
    sudo ls -alg $REDIS_CONF_Ubuntu

    sudo sed -i -e "s/^supervised auto$/supervised systemd/g" $REDIS_CONF_Ubuntu
    sudo egrep -v "(^#|^$)" $REDIS_CONF_Ubuntu |grep "supervised "

    sudo sed -i -e "s/^appendonly no$/appendonly yes/g" $REDIS_CONF_Ubuntu
    sudo egrep -v "(^#|^$)" $REDIS_CONF_Ubuntu |egrep "(appendonly |appendfsync )"

    sudo ls -al /lib/systemd/system/ |grep redis

    sudo systemctl restart redis-server.service
    sudo systemctl status redis-server.service
  else
    echo ""
    echo "This Linux OS ($LinuxOS) is currently not supported and exit"
    exit 1
  fi
elif [ `uname -s` == "Darwin" ]; then
  echo "1. Install and configure Redis on MacOS ......"
  brew --version

  echo ""
  brew install redis
  brew services start redis
  brew services info redis --json

  echo "End to install Redis on MacOS ......"

  echo ""
  echo "2. Enable and Run Redis ......"
  echo "==============================="
  REDIS_CONF_MacOS=/usr/local/etc/redis.conf
  sed -i -e "s/^# supervised auto$/supervised systemd/g" $REDIS_CONF_MacOS
  egrep -v "(^#|^$)" $REDIS_CONF_MacOS |grep "supervised "

  #
  #Configure Redis Persistence using Append Only File (AOF)
  #
  sed -i -e "s/^appendonly no$/appendonly yes/g" $REDIS_CONF_MacOS
  egrep -v "(^#|^$)" $REDIS_CONF_MacOS |egrep "(appendonly |appendfsync )"

  brew services stop redis
  sleep 2
  brew services start redis
  brew services info redis --json
else
  echo ""
  echo "Unknown OS and exit"
  exit 1
fi

echo ""
echo "Sleeping for 5 seconds after Redis installation ......"
sleep 5

echo ""
echo "3. Simply Test Redis ......"
echo "==============================="
which redis-cli
echo "3.1) Test ping ......"
redis-cli ping 

echo ""
echo "3.2) Test write key and value ......"
redis-cli << EOF
SET server:name "fido"
GET server:name
EOF

echo ""
echo "3.3) Test write queue ......"
redis-cli << EOF
lpush demos redis-macOS-demo
rpop demos
EOF

echo ""
echo "Sleep 5 seconds after Redis tests ..."
sleep 5

# 1.Redis Database File (RDB) persistence takes snapshots of the database at intervals corresponding to the save directives in the redis.conf file. The redis.conf file contains three default intervals. RDB persistence generates a compact file for data recovery. However, any writes since the last snapshot is lost.

# 2. Append Only File (AOF) persistence appends every write operation to a log. Redis replays these transactions at startup to restore the database state. You can configure AOF persistence in the redis.conf file with the appendonly and appendfsync directives. This method is more durable and results in less data loss. Redis frequently rewrites the file so it is more concise, but AOF persistence results in larger files, and it is typically slower than the RDB approach

echo ""
echo "************************************************************"
echo "*                                                          *"
echo "* You are successful to install and configure Redis Server *"
echo "*                                                          *"
echo "************************************************************"

exit 0


