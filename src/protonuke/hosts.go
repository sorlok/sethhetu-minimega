package main

import (
	"fmt"
	log "minilog"
	"regexp"
	"strconv"
	"strings"
)

// parse hosts from the command line and return as a map[string]string, where
// the key is the hostname/ip we'll actually target, and the value is the
// parameter provided by the user. This way, if the user provides something
// like 10.0.0.0/24, we can populate the map with 254 keys, all with the same
// value, so we can pretty print things related to the user's input.
func parseHosts(input []string) (map[string]string, error) {
	log.Debugln("parseHosts")
	ret := make(map[string]string)

	for _, i := range input {
		// input can be either a hostname/ip, a subnet, or a comma separated list of the two
		log.Debugln("parsing ", i)

		if strings.Contains(i, ",") { // recursion on comma lists
			d := strings.Split(i, ",")
			log.Debugln("comma delimited: ", d)
			o, err := parseHosts(d)
			if err != nil {
				return nil, err
			} else {
				for k, v := range o {
					ret[k] = v
				}
			}
		} else if strings.Contains(i, "/") { // a subnet
			d := strings.Split(i, "/")
			log.Debugln("subnet ", d)
			if len(d) != 2 {
				return nil, fmt.Errorf("cannot parse %v", i)
			}
			if !isIPv4(d[0]) {
				return nil, fmt.Errorf("network %v is invalid", d[0])
			}
			network := toInt32(d[0])
			cidr, err := strconv.Atoi(d[1])
			if err != nil {
				return nil, err
			}
			if cidr < 0 || cidr > 32 {
				return nil, fmt.Errorf("invalid subnet %v", cidr)
			}

			// we have a valid network and cidr, populate the map
			count := 1 << uint32(32-cidr)
			log.Debug("cidr %v gives %v addresses", cidr, count)
			ip := network

			// special case - if the cidr is < 31, then we remove the network and broadcast address from our calculation
			if cidr < 31 {
				ip++
				count -= 2
			}

			for j := 0; j < count; j++ {
				strIPv4 := toIPv4(ip)
				log.Debug("adding key:value %v:%v", strIPv4, i)
				ret[strIPv4] = i
				ip++
			}
		} else { // host or ip
			if !isIPv4(i) && !isValidDNS(i) {
				return nil, fmt.Errorf("invalid host or ip %v", i)
			}

			log.Debug("adding key:value %v:%v", i, i)
			ret[i] = i
		}
	}

	return ret, nil
}

func isIPv4(ip string) bool {
	d := strings.Split(ip, ".")
	if len(d) != 4 {
		return false
	}

	for _, v := range d {
		octet, err := strconv.Atoi(v)
		if err != nil {
			return false
		}
		if octet < 0 || octet > 255 {
			return false
		}
	}

	return true
}

func toInt32(ip string) uint32 {
	d := strings.Split(ip, ".")

	var ret uint32
	for _, v := range d {
		octet, err := strconv.Atoi(v)
		if err != nil {
			return 0
		}

		ret <<= 8
		ret |= uint32(octet) & 0x000000ff
	}
	return ret
}

func toIPv4(ip uint32) string {
	o0 := (ip & 0xff000000) >> 24
	o1 := (ip & 0x00ff0000) >> 16
	o2 := (ip & 0x0000ff00) >> 8
	o3 := (ip & 0x000000ff)
	return fmt.Sprintf("%v.%v.%v.%v", o0, o1, o2, o3)
}

func isValidDNS(host string) bool {
	// rfc 1123
	expr := `^[[:alnum:]]+[[:alnum:].-]*$`
	matched, err := regexp.MatchString(expr, host)
	if err != nil {
		log.Errorln(err)
		return false
	}
	return matched
}
