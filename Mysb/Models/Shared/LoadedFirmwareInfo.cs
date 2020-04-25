namespace Mysb.Models.Shared
{
    /// <summary>
    /// 
    /// </summary>
    public class LoadedFirmwareInfo
    {
        public ushort Type { get; set; }
        public ushort Version { get; set; }
        public string Path { get; set; } = string.Empty;

    }
}
