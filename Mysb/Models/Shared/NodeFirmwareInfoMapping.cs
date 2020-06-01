using System;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// The shared key info => slug mapping across the application
    /// </summary>
    public class NodeFirmwareInfoMapping
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string NodeId { get; set; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Type { get; set; } = 1;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Version { get; set; } = 1;

        /// <inheritdoc />
        public override string ToString() => $"NodeId: {this.NodeId}, Type: {this.Type}, Version: {this.Version}";
    }
}
