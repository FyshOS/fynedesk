package pulseaudio

import (
	"io"
)

type Server struct {
	PackageName    string
	PackageVersion string
	User           string
	Hostname       string
	SampleSpec     sampleSpec
	DefaultSink    string
	DefaultSource  string
	Cookie         uint32
	ChannelMap     channelMap
}

func (s *Server) ReadFrom(r io.Reader) (int64, error) {
	return 0, bread(r,
		stringTag, &s.PackageName,
		stringTag, &s.PackageVersion,
		stringTag, &s.User,
		stringTag, &s.Hostname,
		&s.SampleSpec,
		stringTag, &s.DefaultSink,
		stringTag, &s.DefaultSource,
		uint32Tag, &s.Cookie,
		&s.ChannelMap)
}

type sink struct {
	Index              uint32
	Name               string
	Description        string
	SampleSpec         sampleSpec
	ChannelMap         channelMap
	ModuleIndex        uint32
	Cvolume            cvolume
	Muted              bool
	MonitorSourceIndex uint32
	MonitorSourceName  string
	Latency            uint64
	Driver             string
	Flags              uint32
	PropList           map[string]string
	RequestedLatency   uint64
	BaseVolume         uint32
	SinkState          uint32
	NVolumeSteps       uint32
	CardIndex          uint32
	Ports              []sinkPort
	ActivePortName     string
	Formats            []formatInfo
}

func (s *sink) ReadFrom(r io.Reader) (int64, error) {
	var portCount uint32
	err := bread(r,
		uint32Tag, &s.Index,
		stringTag, &s.Name,
		stringTag, &s.Description,
		&s.SampleSpec,
		&s.ChannelMap,
		uint32Tag, &s.ModuleIndex,
		&s.Cvolume,
		&s.Muted,
		uint32Tag, &s.MonitorSourceIndex,
		stringTag, &s.MonitorSourceName,
		usecTag, &s.Latency,
		stringTag, &s.Driver,
		uint32Tag, &s.Flags,
		&s.PropList,
		usecTag, &s.RequestedLatency,
		volumeTag, &s.BaseVolume,
		uint32Tag, &s.SinkState,
		uint32Tag, &s.NVolumeSteps,
		uint32Tag, &s.CardIndex,
		uint32Tag, &portCount)
	if err != nil {
		return 0, err
	}
	s.Ports = make([]sinkPort, portCount)
	for i := uint32(0); i < portCount; i++ {
		err = bread(r, &s.Ports[i])
		if err != nil {
			return 0, err
		}
	}
	if portCount == 0 {
		err = bread(r, stringNullTag)
		if err != nil {
			return 0, err
		}
	} else {
		err = bread(r, stringTag, &s.ActivePortName)
		if err != nil {
			return 0, err
		}
	}

	var formatCount uint8
	err = bread(r,
		uint8Tag, &formatCount)
	if err != nil {
		return 0, err
	}
	s.Formats = make([]formatInfo, formatCount)
	for i := uint8(0); i < formatCount; i++ {
		err = bread(r, &s.Formats[i])
		if err != nil {
			return 0, err
		}
	}
	return 0, nil
}

type formatInfo struct {
	Encoding byte
	PropList map[string]string
}

func (i *formatInfo) ReadFrom(r io.Reader) (int64, error) {
	return 0, bread(r, formatInfoTag, uint8Tag, &i.Encoding, &i.PropList)
}

type sinkPort struct {
	Name, Description string
	Pririty           uint32
	Available         uint32
}

func (p *sinkPort) ReadFrom(r io.Reader) (int64, error) {
	return 0, bread(r,
		stringTag, &p.Name,
		stringTag, &p.Description,
		uint32Tag, &p.Pririty,
		uint32Tag, &p.Available)
}

type cvolume []uint32

func (v *cvolume) ReadFrom(r io.Reader) (int64, error) {
	var n byte
	err := bread(r, cvolumeTag, &n)
	if err != nil {
		return 0, err
	}
	*v = make([]uint32, n)
	return 0, bread(r, []uint32(*v))
}

type channelMap []byte

func (m *channelMap) ReadFrom(r io.Reader) (int64, error) {
	var n byte
	err := bread(r, channelMapTag, &n)
	if err != nil {
		return 0, err
	}
	*m = make([]byte, n)
	_, err = r.Read(*m)
	return 0, err
}

type sampleSpec struct {
	Format   byte
	Channels byte
	Rate     uint32
}

func (s *sampleSpec) ReadFrom(r io.Reader) (int64, error) {
	return 0, bread(r, sampleSpecTag, &s.Format, &s.Channels, &s.Rate)
}

type Card struct {
	Index         uint32
	Name          string
	Module        uint32
	Driver        string
	Profiles      map[string]*profile
	ActiveProfile *profile
	PropList      map[string]string
	Ports         []port
}

type profile struct {
	Name, Description string
	Nsinks, Nsources  uint32
	Priority          uint32
	Available         uint32
}

type port struct {
	Card              *Card
	Name, Description string
	Pririty           uint32
	Available         uint32
	Direction         byte
	PropList          map[string]string
	Profiles          []*profile
	LatencyOffset     int64
}

func (p *port) ReadFrom(r io.Reader) (int64, error) {
	err := bread(r,
		stringTag, &p.Name,
		stringTag, &p.Description,
		uint32Tag, &p.Pririty,
		uint32Tag, &p.Available,
		uint8Tag, &p.Direction,
		&p.PropList)
	if err != nil {
		return 0, err
	}
	var portProfileCount uint32
	err = bread(r, uint32Tag, &portProfileCount)
	if err != nil {
		return 0, err
	}
	for j := uint32(0); j < portProfileCount; j++ {
		var profileName string
		err = bread(r, stringTag, &profileName)
		if err != nil {
			return 0, err
		}
		p.Profiles = append(p.Profiles, p.Card.Profiles[profileName])
	}
	return 0, bread(r, int64Tag, &p.LatencyOffset)
}

func (c *Client) sinks() ([]sink, error) {
	b, err := c.request(commandGetSinkInfoList)
	if err != nil {
		return nil, err
	}
	var sinks []sink
	for b.Len() > 0 {
		var sink sink
		err = bread(b, &sink)
		if err != nil {
			return nil, err
		}
		sinks = append(sinks, sink)
	}
	return sinks, nil
}

func (c *Client) Cards() ([]Card, error) {
	b, err := c.request(commandGetCardInfoList)
	if err != nil {
		return nil, err
	}
	var cards []Card
	for b.Len() > 0 {
		var card Card
		var profileCount uint32
		err := bread(b,
			uint32Tag, &card.Index,
			stringTag, &card.Name,
			uint32Tag, &card.Module,
			stringTag, &card.Driver,
			uint32Tag, &profileCount)
		if err != nil {
			return nil, err
		}
		card.Profiles = make(map[string]*profile)
		for i := uint32(0); i < profileCount; i++ {
			var profile profile
			err = bread(b,
				stringTag, &profile.Name,
				stringTag, &profile.Description,
				uint32Tag, &profile.Nsinks,
				uint32Tag, &profile.Nsources,
				uint32Tag, &profile.Priority,
				uint32Tag, &profile.Available)
			if err != nil {
				return nil, err
			}
			card.Profiles[profile.Name] = &profile
		}
		var portCount uint32
		var activeProfileName string
		err = bread(b,
			stringTag, &activeProfileName,
			&card.PropList,
			uint32Tag, &portCount)
		if err != nil {
			return nil, err
		}
		card.ActiveProfile = card.Profiles[activeProfileName]
		card.Ports = make([]port, portCount)
		for i := uint32(0); i < portCount; i++ {
			card.Ports[i].Card = &card
			err = bread(b, &card.Ports[i])
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (c *Client) SetCardProfile(cardIndex uint32, profileName string) error {
	_, err := c.request(commandSetCardProfile,
		uint32Tag, cardIndex,
		stringNullTag,
		stringTag, []byte(profileName), byte(0))
	return err
}

func (c *Client) setDefaultSink(sinkName string) error {
	_, err := c.request(commandSetDefaultSink,
		stringTag, []byte(sinkName), byte(0))
	return err
}

func (c *Client) ServerInfo() (*Server, error) {
	r, err := c.request(commandGetServerInfo)
	if err != nil {
		return nil, err
	}
	var s Server
	err = bread(r, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
