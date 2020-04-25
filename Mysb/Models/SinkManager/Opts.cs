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
        /// <value></value>
        public string SubTopic { get; set; } = "mysensors_rx";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string PubTopic { get; set; } = "mysensors_tx";

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
