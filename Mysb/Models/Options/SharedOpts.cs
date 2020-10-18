using System.Collections.Generic;
using Mysb.Models.Shared;
using TwoMQTT.Core.Interfaces;

namespace Mysb.Models.Options
{
    /// <summary>
    /// The shared options across the application
    /// </summary>
    public record SharedOpts : ISharedOpts<NodeFirmwareInfoMapping>
    {
        public const string Section = "Mysb";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool AutoIDEnabled { get; init; } = false;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort NextID { get; init; } = 1;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string FirmwareBasePath { get; init; } = "/config/firmware";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string SubTopic { get; init; } = "mysensors_rx";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string PubTopic { get; init; } = "mysensors_tx";

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="NodeFirmwareInfoMapping"></typeparam>
        /// <returns></returns>
        public List<NodeFirmwareInfoMapping> Resources { get; init; } = new List<NodeFirmwareInfoMapping>();
    }
}
