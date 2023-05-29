#!/bin/bash

red='\033[0;31m'
green='\033[0;32m'
yellow='\033[0;33m'
plain='\033[0m'

# check root
[[ $EUID -ne 0 ]] && echo -e "${red}Error: ${plain}You must run this script as root!\n" && exit 1

# check config file (/etc/Aiko-Server/aiko.yml) và trong config file có chứa dòng "country_restriction: true" hay không
if [ ! -f "/etc/Aiko-Server/aiko.yml" ]; then
    echo -e "${red}Error: ${plain}Aiko-Server is not installed!\n" && exit 1
else
    if grep -q "country_restriction: true" "/etc/Aiko-Server/aiko.yml"; then
        echo -e "${green}Country Restriction is enabled!${plain}\n"
    else
        echo -e "${red}Error: ${plain}Country Restriction is not enabled!\n" && exit 1
    fi
fi

echo "
*filter
:ufw6-user-input - [0:0]
:ufw6-user-output - [0:0]
:ufw6-user-forward - [0:0]
:ufw6-before-logging-input - [0:0]
:ufw6-before-logging-output - [0:0]
:ufw6-before-logging-forward - [0:0]
:ufw6-user-logging-input - [0:0]
:ufw6-user-logging-output - [0:0]
:ufw6-user-logging-forward - [0:0]
:ufw6-after-logging-input - [0:0]
:ufw6-after-logging-output - [0:0]
:ufw6-after-logging-forward - [0:0]
:ufw6-logging-deny - [0:0]
:ufw6-logging-allow - [0:0]
:ufw6-user-limit - [0:0]
:ufw6-user-limit-accept - [0:0]
### RULES ###
" >> /etc/ufw/user.rules

echo "
*filter
:ufw6-user-input - [0:0]
:ufw6-user-output - [0:0]
:ufw6-user-forward - [0:0]
:ufw6-before-logging-input - [0:0]
:ufw6-before-logging-output - [0:0]
:ufw6-before-logging-forward - [0:0]
:ufw6-user-logging-input - [0:0]
:ufw6-user-logging-output - [0:0]
:ufw6-user-logging-forward - [0:0]
:ufw6-after-logging-input - [0:0]
:ufw6-after-logging-output - [0:0]
:ufw6-after-logging-forward - [0:0]
:ufw6-logging-deny - [0:0]
:ufw6-logging-allow - [0:0]
:ufw6-user-limit - [0:0]
:ufw6-user-limit-accept - [0:0]
### RULES ###
" >> /etc/ufw/user6.rules


# Loop through each country in CountryList
for country in "${CountryList[@]}"
do
    # Download IPv4 and IPv6 lists for the country
    wget -O /tmp/ip_list.txt "https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/${country}/ipv4.txt"
    wget -O /tmp/ip_listv6.txt "https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/${country}/ipv6.txt"

    # Add IPs to UFW rules
    while read ip; do
        # Check if ip is in IpOtherList
        if [[ " ${IpOtherList[@]} " =~ " ${ip} " ]]; then
            echo -e "${red}Error: ${plain}IP ${ip} is in IpOtherList and cannot be added to UFW rules!\n"
        else
            echo "-A ufw-user-input -p tcp -m multiport --dports 22,80,443 -s $ip -j ACCEPT" >> /etc/ufw/user.rules
        fi
    done < /tmp/ip_list.txt

    while read ip; do
        # Check if ip is in IpOtherList
        if [[ " ${IpOtherList[@]} " =~ " ${ip} " ]]; then
            echo -e "${red}Error: ${plain}IP ${ip} is in IpOtherList and cannot be added to UFW rules!\n"
        else
            echo "-A ufw-user-input -p tcp -m multiport --dports 22,80,443 -s $ip -j ACCEPT" >> /etc/ufw/user6.rules
        fi
    done < /tmp/ip_listv6.txt

    # Clean up
    rm /tmp/ip_list.txt
    rm /tmp/ip_listv6.txt
done

# Clean up
rm /tmp/ip_list.txt
rm /tmp/ip_listv6.txt

# Add default UFW rules
echo "
### END RULES ###

### LOGGING ###
-A ufw6-after-logging-input -j LOG --log-prefix "[UFW BLOCK] " -m limit --limit 3/min --limit-burst 10
-A ufw6-after-logging-forward -j LOG --log-prefix "[UFW BLOCK] " -m limit --limit 3/min --limit-burst 10
-I ufw6-logging-deny -m conntrack --ctstate INVALID -j RETURN -m limit --limit 3/min --limit-burst 10
-A ufw6-logging-deny -j LOG --log-prefix "[UFW BLOCK] " -m limit --limit 3/min --limit-burst 10
-A ufw6-logging-allow -j LOG --log-prefix "[UFW ALLOW] " -m limit --limit 3/min --limit-burst 10
### END LOGGING ###

### RATE LIMITING ###
-A ufw6-user-limit -m limit --limit 3/minute -j LOG --log-prefix "[UFW LIMIT BLOCK] "
-A ufw6-user-limit -j REJECT
-A ufw6-user-limit-accept -j ACCEPT
### END RATE LIMITING ###
COMMIT
"


# Reload UFW rules
ufw reload
