using TwoMQTT.Core.Models;

namespace Mysb.Models.SinkManager
{
    /// <summary>
    /// The sink options
    /// </summary>
    public class Opts : MQTTManagerOptions
    {
        public const string Section = "Mysb:MQTT";

        /// <summary>
        /// 
        /// </summary>
        public Opts()
        {
            this.DiscoveryEnabled = false;
            this.TopicPrefix = "home/mysb";
            this.DiscoveryName = "mysb";
        }
    }
}
