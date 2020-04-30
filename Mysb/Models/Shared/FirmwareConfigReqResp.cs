namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public struct FirmwareConfigReqResp
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
        public ushort Blocks { get; set; }

        /// <summary>
        /// 
        /// </summary>
        /// <value></value>
        public ushort Crc { get; set; }

        /// <inheritdoc />
        public override string ToString() =>
            $"Type: {this.Type.ToString()}, Version: {this.Version.ToString()}, " +
            $"Blocks: {this.Blocks.ToString()}, Crc: {this.Crc.ToString()}";
    }
}
