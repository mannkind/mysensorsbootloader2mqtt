using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Runtime.InteropServices;
using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Logging;
using IntelHexFormatReader;
using IntelHexFormatReader.Model;
using Mysb.Models.Shared;

namespace Mysb.DataAccess
{
    public interface IFirmwareDAO
    {
        /// <summary>
        /// 
        /// </summary>
        /// <param name="topic"></param>
        /// <param name="payload"></param>
        /// <returns></returns>
        (string, string) BootloaderCommand(string topic, string payload);

        /// <summary>
        /// Generate a response to a firmware configuration request.
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        Task<string> FirmwareConfigAsync(string nodeId, string payload, CancellationToken cancellationToken = default);

        /// <summary>
        /// Genereate a response to a firmware request.
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="payload"></param>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        Task<string> FirmwareAsync(string nodeId, string payload, CancellationToken cancellationToken = default);
    }

    /// <summary>
    /// 
    /// </summary>
    public class FirmwareDAO : IFirmwareDAO
    {
        /// <summary>
        /// 
        /// </summary>
        /// <param name="logger"></param>
        /// <param name="sharedOpts"></param>
        public FirmwareDAO(ILogger<FirmwareDAO> logger, string firmwareBasePath, IEnumerable<NodeFirmwareInfoMapping> resources)
        {
            this.Logger = logger;
            this.FirmwareBasePath = firmwareBasePath;
            this.Questions = resources;
        }

        /// <inheritdoc />
        public (string, string) BootloaderCommand(string topic, string payload)
        {
            var bootloaderCommand = Const.FirmwareBootloaderCommandTopic.Replace("+/+", string.Empty);
            var partialTopic = topic.Replace(bootloaderCommand, string.Empty);
            var parts = partialTopic.Split('/');
            if (parts.Length != 2)
            {
                this.Logger.LogError("Unable to determine the required parts for a bootloader command");
                return (string.Empty, string.Empty);
            }

            var nodeId = parts[0];
            var cmd = parts[1];
            var type = Convert.ToUInt16(cmd);
            var resp = this.Pack(new FirmwareConfigReqResp
            {
                Type = type,
                Version = type == 0x02 || type == 0x03 ? Convert.ToUInt16(payload) : (ushort)0,
                Blocks = 0,
                Crc = 0xDA7A,
            });

            return (nodeId, resp);
        }

        /// <inheritdoc />
        public async Task<string> FirmwareConfigAsync(string nodeId, string payload, CancellationToken cancellationToken = default)
        {
            var request = this.Unpack<FirmwareConfigReqResp>(payload);
            var fw = this.FirmwareInfo(nodeId, request.Type, request.Version);
            if (fw == null)
            {
                this.Logger.LogError("Firmware Config; From NodeId: {nodeId}; unable to find firmware {type} version {version}", nodeId, request.Type, request.Version);
                return string.Empty;
            }

            var firmware = await this.LoadFromFileAsync(fw.Path, cancellationToken);
            if (firmware == null)
            {
                this.Logger.LogError("Firmware Config; From NodeId: {nodeId}; unable to read firmware {type} version {version}", nodeId, request.Type, request.Version);
                return string.Empty;
            }

            var resp = new FirmwareConfigReqResp
            {
                Type = fw.Type,
                Version = fw.Version,
                Blocks = firmware.Blocks,
                Crc = firmware.Crc,
            };

            this.Logger.LogInformation("FirmmwareConfig Config; NodeId: {nodeId}, {resp}", nodeId, resp);
            return this.Pack(resp);
        }

        /// <inheritdoc />
        public async Task<string> FirmwareAsync(string nodeId, string payload, CancellationToken cancellationToken = default)
        {
            var request = this.Unpack<FirmwareReqResp>(payload);
            var fw = this.FirmwareInfo(nodeId, request.Type, request.Version);
            if (fw == null)
            {
                this.Logger.LogError("Firmware Request; From NodeId: {nodeId}; unable to find firmware {type} version {version}", nodeId, request.Type, request.Version);
                return string.Empty;
            }

            var firmware = await this.LoadFromFileAsync(fw.Path, cancellationToken);
            if (firmware == null)
            {
                this.Logger.LogError("Firmware Request; From NodeId: {nodeId}; unable to read firmware {type} version {version}", nodeId, request.Type, request.Version);
                return string.Empty;
            }

            var resp = new FirmwareReqResp
            {
                Type = fw.Type,
                Version = fw.Version,
                Block = request.Block,
                Data = firmware[request.Block],
            };

            var block = request.Block + 1;
            if (block == firmware.Blocks || block == 1 || block % BLOCK_INTERVAL == 0)
            {
                this.Logger.LogInformation("Firmware Request; From NodeId: {nodeId}, {resp}, Total Blocks: {blocks}", nodeId, resp, firmware.Blocks);
            }

            return this.Pack(resp);
        }


        /// <summary>
        /// Load a firmware file from disk.
        /// </summary>
        /// <param name="path"></param>
        /// <returns></returns>
        protected async Task<Firmware?> LoadFromFileAsync(string path, CancellationToken cancellationToken = default)
        {
            if (!File.Exists(path))
            {
                this.Logger.LogError("File does not exist {path}", path);
                return null;
            }

            var data = new List<byte>();
            var start = 0;
            var end = 0;
            using var reader = new StreamReader(path);
            while (!cancellationToken.IsCancellationRequested)
            {
                var line = await reader.ReadLineAsync();
                if (line == null)
                {
                    break;
                }

                var record = HexFileLineParser.ParseLine(line);
                if (record.RecordType != RecordType.Data)
                {
                    continue;
                }

                if (start == 0 && end == 0)
                {
                    start = record.Address;
                    end = record.Address;
                }

                while (record.Address > end)
                {
                    data.Add(255);
                    end += 1;
                }

                data.AddRange(record.Bytes);
                end += record.Bytes.Length;
            }

            var pad = end % 128;
            foreach (var _ in Enumerable.Range(0, 128 - pad))
            {
                data.Add(255);
                end += 1;
            }

            var blocks = ((end - start) / Const.FirmwareBlockSize);
            var crc = 0xFFFF;
            foreach (var b in data)
            {
                crc = (crc ^ (b & 0xFF));
                foreach (var j in Enumerable.Range(0, 8))
                {
                    var a001 = (crc & 1) > 0;
                    crc = (crc >> 1);
                    if (a001)
                    {
                        crc = (crc ^ 0xA001);
                    }
                }
            }

            return new Firmware
            {
                Blocks = (ushort)blocks,
                Crc = (ushort)crc,
                Data = data,
            };
        }

        /// <summary>
        /// Pack a struct into a hex-encoded string.
        /// </summary>
        /// <param name="obj"></param>
        /// <typeparam name="T"></typeparam>
        /// <returns></returns>
        protected string Pack<T>(T obj)
            where T : struct
        {
            var len = Marshal.SizeOf(obj);
            var b = new byte[len];
            var ptr = Marshal.AllocHGlobal(len);

            Marshal.StructureToPtr(obj, ptr, true);
            Marshal.Copy(ptr, b, 0, len);
            Marshal.FreeHGlobal(ptr);

            return BitConverter.ToString(b).Replace("-", string.Empty);
        }

        /// <summary>
        /// Unpack a hex-encoded string into a struct.
        /// </summary>
        protected T Unpack<T>(string input)
            where T : struct
        {
            var fromBase = 16;
            var byteLen = 2;
            var bytes = new byte[input.Length / byteLen];
            for (var i = 0; i < bytes.Length; i += 1)
            {
                bytes[i] = Convert.ToByte(input.Substring(i * 2, byteLen), fromBase);
            }

            var handle = GCHandle.Alloc(bytes, GCHandleType.Pinned);
            try
            {
                var result = Marshal.PtrToStructure(handle.AddrOfPinnedObject(), typeof(T));
                return result != null ? (T)result : new T();
            }
            finally
            {
                handle.Free();
            }
        }

        /// <summary>
        /// 
        /// </summary>
        private readonly ILogger<FirmwareDAO> Logger;

        /// <summary>
        /// 
        /// </summary>
        /// <typeparam name="NodeMapping"></typeparam>
        /// <returns></returns>
        private readonly IEnumerable<NodeFirmwareInfoMapping> Questions;

        /// <summary>
        /// 
        /// </summary>
        private readonly string FirmwareBasePath;

        /// <summary>
        /// Load firmware information.
        /// In this order, attempt to load the firmware:
        /// * Load the user-defined firmware 
        /// * Load the node-defined firmware
        /// * Load the user-defined default firmware
        /// 
        /// Return null if a firmware was not found.
        /// </summary>
        /// <param name="nodeId"></param>
        /// <param name="firmwareType"></param>
        /// <param name="firmwareVersion"></param>
        /// <returns></returns>
        private LoadedFirmwareInfo? FirmwareInfo(string nodeId, ushort firmwareType, ushort firmwareVersion)
        {
            // Load the user-defined firmware 
            this.Logger.LogDebug("Loading user-defined firmware for nodeId {nodeId}", nodeId);
            var fw = this.Questions.FirstOrDefault(x => x.NodeId == nodeId);
            var path = this.PathToFirmware(fw?.Type, fw?.Version);

            // Load the node-defined firmware
            if (fw == null || !File.Exists(path))
            {
                this.Logger.LogDebug("Loading node-defined firmware for type {firmwareType} and version {firmwareVersion}", firmwareType, firmwareVersion);
                fw = new NodeFirmwareInfoMapping
                {
                    Type = firmwareType,
                    Version = firmwareVersion,
                };
                path = this.PathToFirmware(firmwareType, firmwareVersion);
            }

            // Load the user-defined default firmware
            if (fw == null || !File.Exists(path))
            {
                this.Logger.LogDebug("Loading user-defined default firmware");
                fw = this.Questions.FirstOrDefault(x => x.NodeId == "default");
                path = this.PathToFirmware(fw?.Type, fw?.Version);
            }

            // Unable to load the firmware
            if (fw == null)
            {
                return null;
            }

            return new LoadedFirmwareInfo
            {
                Type = fw.Type,
                Version = fw.Version,
                Path = $"{this.FirmwareBasePath}/{fw.Type}/{fw.Version}/firmware.hex",
            };
        }

        /// <summary>
        /// 
        /// </summary>
        /// <param name="type"></param>
        /// <param name="version"></param>
        /// <returns></returns>
        private string PathToFirmware(ushort? type, ushort? version) =>
            $"{this.FirmwareBasePath}/{type}/{version}/firmware.hex";

        /// <summary>
        /// 
        /// </summary>
        private const ushort BLOCK_INTERVAL = 25;
    }
}
