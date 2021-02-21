using System;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// The shared key info => slug mapping across the application
    /// </summary>
    public record NodeFirmwareInfoMapping
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string NodeId { get; init; } = string.Empty;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Type { get; init; } = 1;

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Version { get; init; } = 1;
    }
}
