using System.Collections.Generic;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// The shared options across the application
    /// </summary>
    public class Opts
    {
        public const string Section = "Mysb";

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public bool AutoIDEnabled { get; set; } = true;

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
        /// <typeparam name="SlugMapping"></typeparam>
        /// <returns></returns>
        public List<NodeFirmwareInfoMapping> Resources { get; set; } = new List<NodeFirmwareInfoMapping>
        {
            new NodeFirmwareInfoMapping
            {
                NodeId = "default",
                Type = 1,
                Version = 1,
            },
        };
    }
}
