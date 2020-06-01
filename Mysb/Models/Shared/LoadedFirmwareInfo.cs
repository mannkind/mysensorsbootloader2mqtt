namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public class LoadedFirmwareInfo
    {
        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Type { get; set; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Version { get; set; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public string Path { get; set; } = string.Empty;

    }
}
