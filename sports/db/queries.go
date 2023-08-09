package db

func getSportEventQuery() string {
	return `
			SELECT 
				id, 
				name, 
				location, 
				advertised_start_time 
			FROM sportevents
		`
}
