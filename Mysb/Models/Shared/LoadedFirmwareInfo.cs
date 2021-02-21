namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public record LoadedFirmwareInfo
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Type { get; init; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Version { get; init; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string Path { get; init; } = string.Empty;
    }
}
