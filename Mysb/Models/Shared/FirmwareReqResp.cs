using System.Runtime.InteropServices;

namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public struct FirmwareReqResp
    {
        public ushort Type { get; set; }
        public ushort Version { get; set; }
        public ushort Block { get; set; }

        [MarshalAs(UnmanagedType.ByValArray, SizeConst = Const.FirmwareBlockSize)]
        public byte[] Data;

        public override string ToString() =>
            $"Type: {this.Type.ToString()}, Version: {this.Version.ToString()}, " +
            $"Block: {this.Block.ToString()}";
    }
}
