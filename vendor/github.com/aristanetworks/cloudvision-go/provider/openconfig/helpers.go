// Copyright (c) 2019 Arista Networks, Inc.
// Use of this source code is governed by the Apache License 2.0
// that can be found in the COPYING file.

package openconfig

// Given an ordered string slice, return the ith element, treating the
// slice as 1-indexed, as SNMP enums are.
func oneIndexed(s []string, i int, def string) string {
	j := i - 1
	if j >= len(s) || j < 0 {
		return def
	}
	return s[j]
}

var interfaceType = []string{
	"other", "regular1822", "hdh1822", "ddnX25",
	"rfc877x25", "ethernetCsmacd", "iso88023Csmacd", "iso88024TokenBus",
	"iso88025TokenRing", "iso88026Man", "starLan", "proteon10Mbit",
	"proteon80Mbit", "hyperchannel", "fddi", "lapb", "sdlc", "ds1",
	"e1", "basicISDN", "primaryISDN", "propPointToPointSerial", "ppp",
	"softwareLoopback", "eon", "ethernet3Mbit", "nsip", "slip", "ultra",
	"ds3", "sip", "frameRelay", "rs232", "para", "arcnet", "arcnetPlus",
	"atm", "miox25", "sonet", "x25ple", "iso88022llc", "localTalk",
	"smdsDxi", "frameRelayService", "v35", "hssi", "hippi", "modem",
	"aal5", "sonetPath", "sonetVT", "smdsIcip", "propVirtual",
	"propMultiplexor", "ieee80212", "fibreChannel", "hippiInterface",
	"frameRelayInterconnect", "aflane8023", "aflane8025", "cctEmul",
	"fastEther", "isdn", "v11", "v36", "g703at64k", "g703at2mb", "qllc",
	"fastEtherFX", "channel", "ieee80211", "ibm370parChan", "escon",
	"dlsw", "isdns", "isdnu", "lapd", "ipSwitch", "rsrb", "atmLogical",
	"ds0", "ds0Bundle", "bsc", "async", "cnr", "iso88025Dtr", "eplrs",
	"arap", "propCnls", "hostPad", "termPad", "frameRelayMPI", "x213",
	"adsl", "radsl", "sdsl", "vdsl", "iso88025CRFPInt", "myrinet",
	"voiceEM", "voiceFXO", "voiceFXS", "voiceEncap", "voiceOverIp",
	"atmDxi", "atmFuni", "atmIma", "pppMultilinkBundle", "ipOverCdlc",
	"ipOverClaw", "stackToStack", "virtualIpAddress", "mpc",
	"ipOverAtm", "iso88025Fiber", "tdlc", "gigabitEthernet", "hdlc",
	"lapf", "v37", "x25mlp", "x25huntGroup", "transpHdlc", "interleave",
	"fast", "ip", "docsCableMaclayer", "docsCableDownstream",
	"docsCableUpstream", "a12MppSwitch", "tunnel", "coffee", "ces",
	"atmSubInterface", "l2vlan", "l3ipvlan", "l3ipxvlan",
	"digitalPowerline", "mediaMailOverIp", "dtm", "dcn", "ipForward",
	"msdsl", "ieee1394", "if-gsn", "dvbRccMacLayer", "dvbRccDownstream",
	"dvbRccUpstream", "atmVirtual", "mplsTunnel", "srp", "voiceOverAtm",
	"voiceOverFrameRelay", "idsl", "compositeLink", "ss7SigLink",
	"propWirelessP2P", "frForward", "rfc1483", "usb", "ieee8023adLag",
	"bgppolicyaccounting", "frf16MfrBundle", "h323Gatekeeper",
	"h323Proxy", "mpls", "mfSigLink", "hdsl2", "shdsl", "ds1FDL", "pos",
	"dvbAsiIn", "dvbAsiOut", "plc", "nfas", "tr008", "gr303RDT",
	"gr303IDT", "isup", "propDocsWirelessMaclayer",
	"propDocsWirelessDownstream", "propDocsWirelessUpstream",
	"hiperlan2", "propBWAp2Mp", "sonetOverheadChannel",
	"digitalWrapperOverheadChannel", "aal2", "radioMAC", "atmRadio",
	"imt", "mvl", "reachDSL", "frDlciEndPt", "atmVciEndPt",
	"opticalChannel", "opticalTransport", "propAtm", "voiceOverCable",
	"infiniband", "teLink", "q2931", "virtualTg", "sipTg", "sipSig",
	"docsCableUpstreamChannel", "econet", "pon155", "pon622", "bridge",
	"linegroup", "voiceEMFGD", "voiceFGDEANA", "voiceDID",
	"mpegTransport", "sixToFour", "gtp", "pdnEtherLoop1",
	"pdnEtherLoop2", "opticalChannelGroup", "homepna", "gfp",
	"ciscoISLvlan", "actelisMetaLOOP", "fcipLink", "rpr", "qam", "lmp",
	"cblVectaStar", "docsCableMCmtsDownstream", "adsl2",
	"macSecControlledIF", "macSecUncontrolledIF", "aviciOpticalEther",
	"atmbond", "voiceFGDOS", "mocaVersion1", "ieee80216WMAN",
	"adsl2plus", "dvbRcsMacLayer", "dvbTdm", "dvbRcsTdma", "x86Laps",
	"wwanPP", "wwanPP2", "voiceEBS", "ifPwType", "ilan", "pip",
	"aluELP", "gpon", "vdsl2", "capwapDot11Profile", "capwapDot11Bss",
	"capwapWtpVirtualRadio", "bits", "docsCableUpstreamRfPort",
	"cableDownstreamRfPort", "vmwareVirtualNic", "ieee802154", "otnOdu",
	"otnOtu", "ifVfiType", "g9981", "g9982", "g9983", "aluEpon",
	"aluEponOnu", "aluEponPhysicalUni", "aluEponLogicalLink",
	"aluGponOnu", "aluGponPhysicalUni", "vmwareNicTeam",
}

// InterfaceType returns the SNMP interface type string
// corresponding to an interface type value. OpenConfig uses
// the same types.
func InterfaceType(t int) string {
	return oneIndexed(interfaceType, t, interfaceType[0])
}

var intfAdminStatus = []string{
	"UP",
	"DOWN",
	"TESTING",
}

// IntfAdminStatus returns the SNMP interface admin status type
// string corresponding to the provided value. OpenConfig uses
// the same types.
func IntfAdminStatus(t int) string {
	return oneIndexed(intfAdminStatus, t, "")
}

var intfOperStatus = []string{
	"UP",
	"DOWN",
	"TESTING",
	"UNKNOWN",
	"DORMANT",
	"NOT_PRESENT",
	"LOWER_LAYER_DOWN",
}

// IntfOperStatus returns the SNMP interface oper status type
// string corresponding to the provided value. OpenConfig uses
// the same types.
func IntfOperStatus(t int) string {
	return oneIndexed(intfOperStatus, t, "")
}

var lldpChassisIDType = []string{
	"CHASSIS_COMPONENT",
	"INTERFACE_ALIAS",
	"PORT_COMPONENT",
	"MAC_ADDRESS",
	"NETWORK_ADDRESS",
	"INTERFACE_NAME",
	"LOCAL",
}

// LLDPChassisIDType returns the SNMP LLDP chassis ID type string
// corresponding to the provided value. OpenConfig uses the same
// types.
func LLDPChassisIDType(t int) string {
	return oneIndexed(lldpChassisIDType, t, "")
}

var lldpPortIDType = []string{
	"INTERFACE_ALIAS",
	"PORT_COMPONENT",
	"MAC_ADDRESS",
	"NETWORK_ADDRESS",
	"INTERFACE_NAME",
	"AGENT_CIRCUIT_ID",
	"LOCAL",
}

// LLDPPortIDType returns the SNMP LLDP port ID type string
// corresponding to the provided value. OpenConfig uses the same
// types.
func LLDPPortIDType(t int) string {
	return oneIndexed(lldpPortIDType, t, "")
}
