namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public struct FirmwareConfigReqResp
    {
        public ushort Type { get; set; }
        public ushort Version { get; set; }
        public ushort Blocks { get; set; }
        public ushort Crc { get; set; }

        public override string ToString() =>
            $"Type: {this.Type.ToString()}, Version: {this.Version.ToString()}, " +
            $"Blocks: {this.Blocks.ToString()}, Crc: {this.Crc.ToString()}";
    }
}
