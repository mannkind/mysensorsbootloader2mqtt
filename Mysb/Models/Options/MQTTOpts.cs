using TwoMQTT.Models;

namespace Mysb.Models.Options;

/// <summary>
/// The sink options
/// </summary>
public record MQTTOpts : MQTTManagerOptions
{
    public const string Section = "Mysb:MQTT";

    /// <summary>
    /// 
    /// </summary>
    public MQTTOpts()
    {
        this.DiscoveryEnabled = false;
        this.TopicPrefix = "home/mysb";
        this.DiscoveryName = "mysb";
    }
}
