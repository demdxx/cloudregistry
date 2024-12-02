package etcd

func servicePrefix(name string) string {
	return "services/" + name + "/"
}

func servicePathKey(name, id string) string {
	return "services/" + name + "/" + id
}
