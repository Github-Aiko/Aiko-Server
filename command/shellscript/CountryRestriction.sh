#!/bin/bash

# Set color variables
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
PLAIN='\033[0m'

# Check root
if [[ $EUID -ne 0 ]]; then 
  echo -e "${RED}Error:${PLAIN} You must run this script as root!\n" 
  exit 1 
fi 

# Check config file (/etc/Aiko-Server/aiko.yml) và trong config file có chứa dòng "country_restriction: true" hay không
AIKO_YML="/etc/Aiko-Server/aiko.yml"
if [ ! -f "$AIKO_YML" ] || ! grep --quiet "CountryRestriction: true" "$AIKO_YML"; then 
  echo -e "${RED}Error:${PLAIN} Country Restriction is not enabled!\n" 
  exit 1 
else
  echo -e "${GREEN}Country Restriction is enabled!${PLAIN}\n"
fi 

# Set variables for file paths
USER_RULES="/etc/ufw/user.rules"
USER6_RULES="/etc/ufw/user6.rules"
TMP_IPV4_LIST="/tmp/ip_list.txt"
TMP_IPV6_LIST="/tmp/ip_listv6.txt"

# Clear user rules files
echo "
*filter
:ufw-user-input - [0:0]
:ufw-user-output - [0:0]
:ufw-user-forward - [0:0]
:ufw-before-logging-input - [0:0]
:ufw-before-logging-output - [0:0]
:ufw-before-logging-forward - [0:0]
:ufw-user-logging-input - [0:0]
:ufw-user-logging-output - [0:0]
:ufw-user-logging-forward - [0:0]
:ufw-after-logging-input - [0:0]
:ufw-after-logging-output - [0:0]
:ufw-after-logging-forward - [0:0]
:ufw-logging-deny - [0:0]
:ufw-logging-allow - [0:0]
:ufw-user-limit - [0:0]
:ufw-user-limit-accept - [0:0]
### RULES ###
" > "$USER_RULES"

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
" > "$USER6_RULES"

# Checkletter converts all country names in the given CountryList to lowercase.
# This is done because the country lists are stored in lowercase.
checkletter() {
    # Loop through each country in CountryList
    for country in "${CountryList[@]}"
    do
        # Convert country name to lowercase
        country=$(echo "$country" | tr '[:upper:]' '[:lower:]')

        # Replace country name in CountryList with lowercase version
        CountryList=("${CountryList[@]/$country}")
        CountryList=("${CountryList[@]}" "$country")
    done
}

# Read CountryList and IpOtherList from config file
if [ ! -f "$AIKO_YML" ]; then
    echo -e "${RED}Error:${PLAIN} Aiko-Server is not installed!\n" 
    exit 1
else
    # Read CountryList from config file
    if country_list=$(grep "CountryList" "$AIKO_YML" | awk '{print $2}' | tr -d '[:space:]"' | tr ',' ' '); then
        CountryList=($country_list)
        checkletter
    else
        echo -e "${RED}Error:${PLAIN} CountryList is not defined in $AIKO_YML!\n" 
        exit 1
    fi

    # Read IpOtherList from config file
    if ip_other_list=$(grep "IpOtherList" "$AIKO_YML" | awk '{print $2}' | tr -d '[:space:]"' | tr ',' ' '); then
        IpOtherList=($ip_other_list)
    else
        echo -e "${RED}Error:${PLAIN} IpOtherList is not defined in $AIKO_YML!\n" 
        exit 1
    fi
fi

# Loop through each country in CountryList
for country in "${CountryList[@]}"
do
    # Download IPv4 and IPv6 lists for the country
    curl -sL "https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/${country}/ipv4.txt" -o "$TMP_IPV4_LIST"
    curl -sL "https://raw.githubusercontent.com/Github-Aiko/IPLocation/master/${country}/ipv6.txt" -o "$TMP_IPV6_LIST"

    # Add IPs to UFW rules
    while read -r ip; do
        # Check if ip is in IpOtherList
        if [[ " ${IpOtherList[@]} " =~ " ${ip} " ]]; then
            echo -e "${RED}Error:${PLAIN} IP $ip is in IpOtherList and cannot be added to UFW rules!\n"
        else
            echo "-A ufw-user-input -p tcp -m multiport --dports 22,80,443 -s $ip -j ACCEPT" >> "$USER_RULES"
        fi
    done < "$TMP_IPV4_LIST"

    while read -r ip; do
        # Check if ip is in IpOtherList
        if [[ " ${IpOtherList[@]} " =~ " ${ip} " ]]; then
            echo -e "${RED}Error:${PLAIN} IP $ip is in IpOtherList and cannot be added to UFW rules!\n"
        else
            echo "-A ufw-user-input -p tcp -m multiport --dports 22,80,443 -s $ip -j ACCEPT" >> "$USER6_RULES"
        fi
    done < "$TMP_IPV6_LIST"
done

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
" >> "$USER_RULES"

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
" >> "$USER6_RULES"

# Check if UFW is enabled
if ufw status | grep -q "Status: active"; then
    # Check if UFW rules are already loaded
    if ufw status verbose | grep -q "user.rules"; then
        # Reload UFW rules
        ufw reload
    else
        # Add UFW rules
        ufw --force enable
    fi
fi