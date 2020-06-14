using System.Collections.Generic;
using Mysb.Models.Shared;

namespace Mysb.Models.Options
{
    /// <summary>
    /// The shared options across the application
    /// </summary>
    public class SharedOpts
    {
        public const string Section = "Mysb";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool AutoIDEnabled { get; set; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort NextID { get; set; } = 1;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string FirmwareBasePath { get; set; } = "/config/firmware";

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
        /// <typeparam name="NodeFirmwareInfoMapping"></typeparam>
        /// <returns></returns>
        public List<NodeFirmwareInfoMapping> Resources { get; set; } = new List<NodeFirmwareInfoMapping>();
    }
}
