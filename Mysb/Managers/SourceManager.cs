using System.Threading;
using System.Threading.Tasks;
using Microsoft.Extensions.Hosting;

namespace Mysb.Managers
{
    /// <summary>
    /// An abstract class representing a managed way to poll a source.
    /// </summary>
    public class SourceManager : BackgroundService
    {
        /// <summary>
        /// Executed as an IHostedService as a background job.
        /// </summary>
        /// <param name="cancellationToken"></param>
        /// <returns></returns>
        protected override Task ExecuteAsync(CancellationToken cancellationToken = default) =>
            Task.CompletedTask;
    }
}
