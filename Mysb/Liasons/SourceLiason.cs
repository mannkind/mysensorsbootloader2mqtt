using System;
using System.Collections.Generic;
using System.Linq;
using System.Runtime.CompilerServices;
using System.Threading;
using System.Threading.Tasks;
using TwoMQTT.Core.Interfaces;

namespace Mysb.Liasons
{
    /// <summary>
    /// I'm just here so I won't be fined.
    /// </summary>
    public class SourceLiason : ISourceLiason<object, object>
    {
        /// <inheritdoc />
        public IAsyncEnumerable<object?> FetchAllAsync(CancellationToken cancellationToken = default) =>
            AsyncEnumerable.Empty<object?>();
    }
}