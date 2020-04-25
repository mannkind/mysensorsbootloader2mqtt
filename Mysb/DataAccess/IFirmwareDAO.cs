using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
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
}
