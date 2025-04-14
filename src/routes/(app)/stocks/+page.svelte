<script lang="ts">
	import { onMount } from 'svelte';
	import { formatCurrency, formatDate } from '$lib/utils';
	import { ChevronUp, ChevronDown, Edit2, Check, X, Plus, Tag } from 'lucide-svelte';

	interface Stock {
		symbol: string;
		averagePrice: number;
		lastTradedPrice: number;
		targetPrice: number;
		shares: number;
		totalInvestment: number;
		gainPercent: number;
		gainAmount: number;
		drawdownFromPeak: number;
		lastPurchaseDate: string;
		tags: { tag: string; color: string }[];
	}

	let stocks: Stock[] = [];
	let loading = true;
	let sortColumn: keyof Stock = 'symbol';
	let sortDirection: 'asc' | 'desc' = 'asc';
	let editingTargetPrice: string | null = null;
	let newTargetPrice: number = 0;
	let ltpLoading: { [key: string]: boolean } = {};
	let editingTag: { symbol: string; tag: string } | null = null;
	let newTag: string = '';
	let newTagColor: string = '#4F46E5';

	onMount(async () => {
		try {
			const response = await fetch('/api/stocks');
			const data = await response.json();
			stocks = data.stocks;
		} catch (error) {
			console.error('Error fetching stocks:', error);
		} finally {
			loading = false;
		}
	});

	async function updateTargetPrice(symbol: string) {
		try {
			const response = await fetch('/api/stocks/target-price', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ symbol, targetPrice: newTargetPrice }),
			});
			if (response.ok) {
				const updatedStock = await response.json();
				stocks = stocks.map(stock => 
					stock.symbol === symbol ? { ...stock, targetPrice: updatedStock.targetPrice } : stock
				);
				editingTargetPrice = null;
			}
		} catch (error) {
			console.error('Error updating target price:', error);
		}
	}

	function startEditing(symbol: string, currentPrice: number) {
		editingTargetPrice = symbol;
		newTargetPrice = currentPrice;
	}

	function cancelEditing() {
		editingTargetPrice = null;
	}

	async function addTag(symbol: string) {
		try {
			const response = await fetch('/api/stocks/tag', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ symbol, tag: newTag, color: newTagColor }),
			});
			if (response.ok) {
				const updatedStocks = await fetch('/api/stocks').then(r => r.json());
				stocks = updatedStocks.stocks;
				editingTag = null;
				newTag = '';
			}
		} catch (error) {
			console.error('Error adding tag:', error);
		}
	}

	async function removeTag(symbol: string, tag: string) {
		try {
			const response = await fetch('/api/stocks/tag', {
				method: 'DELETE',
				headers: {
					'Content-Type': 'application/json',
				},
				body: JSON.stringify({ symbol, tag }),
			});
			if (response.ok) {
				const updatedStocks = await fetch('/api/stocks').then(r => r.json());
				stocks = updatedStocks.stocks;
			}
		} catch (error) {
			console.error('Error removing tag:', error);
		}
	}

	function startAddingTag(symbol: string) {
		editingTag = { symbol, tag: '' };
	}

	function cancelAddingTag() {
		editingTag = null;
		newTag = '';
	}

	$: sortedStocks = [...stocks].sort((a, b) => {
		let comparison = 0;
		if (sortColumn === 'symbol' || sortColumn === 'lastPurchaseDate') {
			comparison = String(a[sortColumn]).localeCompare(String(b[sortColumn]));
		} else {
			comparison = Number(a[sortColumn]) - Number(b[sortColumn]);
		}
		return sortDirection === 'asc' ? comparison : -comparison;
	});

	function sortStocks(column: keyof Stock) {
		if (sortColumn === column) {
			sortDirection = sortDirection === 'asc' ? 'desc' : 'asc';
		} else {
			sortColumn = column;
			sortDirection = 'asc';
		}
	}

	function getSortIcon(column: keyof Stock) {
		if (sortColumn !== column) return null;
		return sortDirection === 'asc' ? ChevronUp : ChevronDown;
	}
</script>

<div class="container mx-auto px-2 py-8 max-w-[95%]">
	<h1 class="text-2xl font-bold mb-6">Stocks Portfolio</h1>

	{#if loading}
		<div class="flex justify-center items-center h-64">
			<div class="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
		</div>
	{:else}
		<div class="overflow-x-auto">
			<table class="min-w-full bg-white rounded-lg overflow-hidden text-xs">
				<thead class="bg-gray-50">
					<tr>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('symbol')}
						>
							<div class="flex items-center gap-1">
								Symbol
								<svelte:component this={sortColumn === 'symbol' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('shares')}
						>
							<div class="flex items-center gap-1">
								Units
								<svelte:component this={sortColumn === 'shares' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('averagePrice')}
						>
							<div class="flex items-center gap-1">
								Avg Price
								<svelte:component this={sortColumn === 'averagePrice' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('lastTradedPrice')}
						>
							<div class="flex items-center gap-1">
								LTP
								<svelte:component this={sortColumn === 'lastTradedPrice' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('totalInvestment')}
						>
							<div class="flex items-center gap-1">
								Invested
								<svelte:component this={sortColumn === 'totalInvestment' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('lastTradedPrice')}
						>
							<div class="flex items-center gap-1">
								Current
								<svelte:component this={sortColumn === 'lastTradedPrice' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('gainAmount')}
						>
							<div class="flex items-center gap-1">
								Net Gain
								<svelte:component this={sortColumn === 'gainAmount' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('gainPercent')}
						>
							<div class="flex items-center gap-1">
								Gain%
								<svelte:component this={sortColumn === 'gainPercent' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('drawdownFromPeak')}
						>
							<div class="flex items-center gap-1">
								Drawdown
								<svelte:component this={sortColumn === 'drawdownFromPeak' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm"
							on:click={() => sortStocks('lastPurchaseDate')}
						>
							<div class="flex items-center gap-1">
								Last purchased at
								<svelte:component this={sortColumn === 'lastPurchaseDate' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
						<th
							class="px-3 py-3 text-left font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100 text-sm border-l-2 border-gray-300"
							on:click={() => sortStocks('targetPrice')}
						>
							<div class="flex items-center gap-1">
								Target Price
								<svelte:component this={sortColumn === 'targetPrice' ? (sortDirection === 'asc' ? ChevronUp : ChevronDown) : null} size={14} />
							</div>
						</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-200">
					{#each sortedStocks as stock}
						<tr class="hover:bg-gray-50">
							<td class="px-3 py-4 whitespace-nowrap font-medium text-gray-900 text-base">
								<div class="flex items-center gap-2">
									{stock.symbol}
									<div class="flex flex-wrap gap-1">
										{#each stock.tags || [] as tag}
											<span
												class="px-2 py-1 text-xs rounded-full flex items-center gap-1"
												style="background-color: {tag.color}20; color: {tag.color}"
											>
												{tag.tag}
												<button
													on:click={() => removeTag(stock.symbol, tag.tag)}
													class="hover:text-gray-700"
												>
													<X size={12} />
												</button>
											</span>
										{/each}
										{#if editingTag?.symbol === stock.symbol}
											<div class="flex items-center gap-1">
												<input
													type="text"
													bind:value={newTag}
													class="w-24 px-2 py-1 border rounded text-sm"
													placeholder="New tag"
												/>
												<input
													type="color"
													bind:value={newTagColor}
													class="w-6 h-6 rounded cursor-pointer"
												/>
												<button
													on:click={() => addTag(stock.symbol)}
													class="text-green-600 hover:text-green-800"
												>
													<Check size={16} />
												</button>
												<button
													on:click={cancelAddingTag}
													class="text-red-600 hover:text-red-800"
												>
													<X size={16} />
												</button>
											</div>
										{:else}
											<button
												on:click={() => startAddingTag(stock.symbol)}
												class="p-0.5 text-gray-400 hover:text-gray-600 hover:bg-gray-100 rounded transition-colors"
												title="Add tag"
											>
												<Plus size={12} />
											</button>
										{/if}
									</div>
								</div>
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{stock.shares}
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{formatCurrency(stock.averagePrice, 2)}
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{formatCurrency(stock.lastTradedPrice, 2)}
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{formatCurrency(stock.totalInvestment, 2)}
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{formatCurrency(stock.lastTradedPrice * stock.shares, 2)}
							</td>
							<td
								class="px-3 py-4 whitespace-nowrap {stock.gainAmount >= 0
									? 'text-green-600'
									: 'text-red-600'} text-base"
							>
								{formatCurrency(stock.gainAmount, 2)}
							</td>
							<td
								class="px-3 py-4 whitespace-nowrap {stock.gainPercent >= 0
									? 'text-green-600'
									: 'text-red-600'} text-base"
							>
								{stock.gainPercent.toFixed(2)}%
							</td>
							<td
								class="px-3 py-4 whitespace-nowrap {stock.drawdownFromPeak >= 0
									? 'text-green-600'
									: 'text-red-600'} text-base"
							>
								{(stock.drawdownFromPeak || 0).toFixed(2)}%
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base">
								{formatDate(stock.lastPurchaseDate)}
							</td>
							<td class="px-3 py-4 whitespace-nowrap text-gray-500 text-base border-l-2 border-gray-300">
								{#if editingTargetPrice === stock.symbol}
									<div class="flex items-center gap-2">
										<input
											type="number"
											bind:value={newTargetPrice}
											class="w-24 px-2 py-1 border rounded"
											step="0.01"
											min="0"
										/>
										<button
											on:click={() => updateTargetPrice(stock.symbol)}
											class="text-green-600 hover:text-green-800"
										>
											<Check size={16} />
										</button>
										<button
											on:click={cancelEditing}
											class="text-red-600 hover:text-red-800"
										>
											<X size={16} />
										</button>
									</div>
								{:else}
									<div class="flex items-center gap-1">
										{#if stock.targetPrice > 0}
											{formatCurrency(stock.targetPrice, 2)}
										{/if}
										<button
											on:click={() => startEditing(stock.symbol, stock.targetPrice || 0)}
											class="p-0.7 text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition-colors"
											title="Edit target price"
										>
											<Edit2 size={8} />
										</button>
									</div>
								{/if}
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div> 