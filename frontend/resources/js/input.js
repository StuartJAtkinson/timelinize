async function inputPageMain() {
	const qs = new URLSearchParams(window.location.search);
	const repoID = qs.get('repo_id');
	const jobID = Number(qs.get('job_id'));
	const submitButton = $('#submit-graph');
	const skipButton = $('#skip-graph');
	const errorMessage = $('#interactive-error');
	let currentGraph;
	let busy = false;
	let originalTimestamp;
	let originalClassification;

	function setBusy(value) {
		busy = value;
		submitButton.disabled = value || !currentGraph;
		skipButton.disabled = value || !currentGraph;
	}

	function showError(error) {
		errorMessage.innerText = error?.message || String(error);
		errorMessage.classList.remove('d-none');
	}

	function dateTimeLocalValue(timestamp) {
		if (!timestamp) return '';
		const date = new Date(timestamp);
		if (Number.isNaN(date.getTime())) return '';
		const localDate = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
		return localDate.toISOString().slice(0, 16);
	}

	function renderGraph(graph) {
		currentGraph = graph;
		errorMessage.classList.add('d-none');
		const content = $('#item-content');
		content.replaceChildren();

		if (graph.item) {
			const item = graph.item;
			const contentData = item.content || {};
			const fakeItemRow = {
				classification: item.classification?.name || 'Item',
				filename: contentData.filename,
				data_type: contentData.media_type || 'application/octet-stream',
				data_text: contentData.preview_text,
				intermediate_location: item.intermediate_location,
				metadata: item.metadata,
				owner: item.owner,
				timestamp: item.timestamp,
				location: item.location,
				repo_id: repoID,
				interactive: contentData.data_available ? {
					job_id: jobID,
					graph_id: graph.processing_id
				} : undefined
			};
			content.append(itemContentElement(fakeItemRow));
			$('#interactive-item-fields').classList.remove('d-none');
			originalTimestamp = item.timestamp;
			originalClassification = item.classification?.name || '';
			$('#interactive-timestamp').value = dateTimeLocalValue(originalTimestamp);
			$('#interactive-classification').value = originalClassification;
			$('#item-class-label').innerText = item.classification?.name || 'Item';
			$('#item-media-type').innerText = contentData.media_type || '—';
			$('#item-original-path').innerText = item.original_location || '—';
			$('#item-intermediate-location').innerText = item.intermediate_location || '—';
			$('#item-coordinate').innerText = item.location?.latitude != null && item.location?.longitude != null
				? `${item.location.latitude}, ${item.location.longitude}`
				: '—';
			$('#item-altitude').innerText = item.location?.altitude ?? '—';
			$('#item-size').innerText = contentData.size ? humanizeBytes(contentData.size) : '—';
			$('#item-stored').innerText = 'Pending review';

			const metadata = $('#item-metadata');
			metadata.replaceChildren();
			for (const [key, value] of Object.entries(item.metadata || {})) {
				const entry = document.createElement('div');
				entry.className = 'datagrid-item';
				const title = document.createElement('div');
				title.className = 'datagrid-title';
				title.innerText = key;
				const text = document.createElement('div');
				text.className = 'datagrid-content';
				text.innerText = typeof value === 'string' ? value : JSON.stringify(value);
				entry.append(title, text);
				metadata.append(entry);
			}
		} else if (graph.entity) {
			const entity = graph.entity;
			const title = document.createElement('h3');
			title.innerText = entity.name || entity.type || 'Entity';
			content.append(title);
			$('#interactive-item-fields').classList.add('d-none');
			$('#item-metadata').replaceChildren();
			$('#item-media-type').innerText = '—';
			$('#item-original-path').innerText = '—';
			$('#item-intermediate-location').innerText = '—';
			$('#item-coordinate').innerText = '—';
			$('#item-altitude').innerText = '—';
			$('#item-size').innerText = '—';
			$('#item-stored').innerText = 'Pending review';
		} else {
			throw new Error('The import produced an empty graph.');
		}
	}

	async function loadNextGraph() {
		setBusy(true);
		errorMessage.classList.add('d-none');
		try {
			const graph = await app.NextGraph(repoID, jobID);
			if (!graph) {
				currentGraph = null;
				$('#item-content').innerText = 'Import complete.';
				$('#interactive-item-fields').classList.add('d-none');
				return;
			}
			renderGraph(graph);
		} catch (error) {
			showError(error);
		} finally {
			setBusy(false);
		}
	}

	async function submitCurrentGraph(skip) {
		if (busy || !currentGraph) return;
		setBusy(true);
		errorMessage.classList.add('d-none');
		try {
			if (!skip && currentGraph.item) {
				const timestamp = $('#interactive-timestamp').value;
				if (timestamp && timestamp !== dateTimeLocalValue(originalTimestamp)) {
					currentGraph.item.timestamp = new Date(timestamp).toISOString();
				}
				const classification = $('#interactive-classification').value.trim();
				if (classification && classification !== originalClassification) {
					currentGraph.item.classification ||= {};
					currentGraph.item.classification.name = classification;
				}
			}
			await app.SubmitGraph(repoID, jobID, currentGraph, skip);
			currentGraph = null;
			await loadNextGraph();
		} catch (error) {
			showError(error);
			setBusy(false);
		}
	}

	submitButton.addEventListener('click', () => submitCurrentGraph(false));
	skipButton.addEventListener('click', () => submitCurrentGraph(true));
	await loadNextGraph();
}
