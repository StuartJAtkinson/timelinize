async function inputPageMain() {
	// const dp = new AirDatepicker($('#item-date'), {

	// });
	// $('#item-date').append(newDatePicker({
	// 	passthru: {
	// 		range: false
	// 	},
	// 	rangeToggle: true,
	// 	time: false,
	// 	timeToggle: false,
	// 	proximity: false,
	// 	vertical: true,
	// 	sort: false
	// }));

	const qs = new URLSearchParams(window.location.search);
	const repoID = qs.get('repo_id');
	const jobID = Number(qs.get('job_id'));

	const graph = await app.NextGraph(repoID, jobID);
	console.log("FIRST GRAPH:", graph);

	const fakeItemRow = {
		classification: graph.item.classification.name,
		filename: graph.item.content.filename,
		data_type: graph.item.content.media_type,
		intermediate_location: graph.item.intermediate_location,
		metadata: graph.item.metadata,
		owner: graph.item.owner,
		timestamp: graph.item.timestamp,
		location: graph.item.location,
		repo_id: repoID,
		interactive: {
			job_id: jobID,
			graph_id: graph.processing_id
		}
	};
	console.log("FAKE ITEM ROW:", fakeItemRow);

	let itemContentEl = itemContentElement(fakeItemRow);
	$('#item-content').append(itemContentEl);

	await app.SubmitGraph(repoID, jobID, graph, false);

	const graph2 = await app.NextGraph(repoID, jobID);
	console.log("NEXT GRAPH:", graph2);
}