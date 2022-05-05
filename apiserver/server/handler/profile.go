package handler

import (
	"context"
	"github.com/emicklei/go-restful/v3"
	"github.com/opensource-f2f/open-podcasts/api/osf2f.my.domain/v1alpha1"
	client "github.com/opensource-f2f/open-podcasts/generated/clientset/versioned"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"net/http"
)

type Profile struct {
	nameQuery    *restful.Parameter
	rssQuery     *restful.Parameter
	episodeQuery *restful.Parameter
	profilePath  *restful.Parameter
	feishuQuery  *restful.Parameter
}

func (r Profile) WebService() (ws *restful.WebService) {
	ws = new(restful.WebService)
	ws.Path("/profiles")

	// set the parameters
	r.nameQuery = restful.QueryParameter("name", "The name of a profile")
	r.rssQuery = restful.QueryParameter("rss", "The name of a rss")
	r.episodeQuery = restful.QueryParameter("episode", "The name of an episode")
	r.profilePath = restful.PathParameter("profile", "The name of a profile")
	r.feishuQuery = restful.QueryParameter("feishu", "The webhook of Feishu")

	// set the routes
	ws.Route(ws.POST("/").
		Param(r.nameQuery.Required(true)).
		To(r.create).
		Returns(http.StatusOK, "OK", []RSS{}))
	ws.Route(ws.GET("/{profile}").
		Param(r.profilePath).
		To(r.findOne).
		Returns(http.StatusOK, "OK", v1alpha1.Profile{}))
	ws.Route(ws.POST("/{profile}/subscribe").
		Param(r.profilePath).
		Param(r.rssQuery).
		To(r.subscribe).
		Returns(http.StatusOK, "OK", []RSS{}))
	ws.Route(ws.POST("/{profile}/unsubscribe").
		Param(r.profilePath).
		Param(r.rssQuery).
		To(r.unsubscribe).
		Returns(http.StatusOK, "OK", []RSS{}))
	ws.Route(ws.GET("/{profile}/subscriptions").
		Param(r.profilePath).
		To(r.subscriptions).
		Returns(http.StatusOK, "OK", []v1alpha1.Subscription{}))
	ws.Route(ws.POST("/{profile}/playLater").
		Param(r.profilePath).
		Param(r.episodeQuery).
		To(r.playLater).
		Returns(http.StatusOK, "OK", []v1alpha1.RSS{}))
	ws.Route(ws.DELETE("/{profile}/playLater").
		Param(r.profilePath).
		Param(r.episodeQuery).
		To(r.playOver).
		Returns(http.StatusOK, "OK", []v1alpha1.RSS{}))
	ws.Route(ws.POST("/{profile}/notifier").
		Param(r.profilePath).
		Param(r.feishuQuery).
		To(r.notifier).
		Returns(http.StatusOK, "OK", []v1alpha1.RSS{}))
	return
}

func (r Profile) create(request *restful.Request, response *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	name := request.QueryParameter(r.nameQuery.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	_, _ = clientset.Osf2fV1alpha1().Profiles(ns).Create(ctx, &v1alpha1.Profile{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}, metav1.CreateOptions{})
	response.Write([]byte("ok"))
}

func (r Profile) findOne(request *restful.Request, response *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	name := request.PathParameter(r.profilePath.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, _ := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, name, metav1.GetOptions{})
	response.WriteAsJson(profile)
}

func (r Profile) subscribe(request *restful.Request, response *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	rss := request.QueryParameter(r.rssQuery.Data().Name)
	profileName := request.PathParameter(r.profilePath.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	if profile.Spec.Subscription.Name == "" {
		sub, _ := clientset.Osf2fV1alpha1().Subscriptions(ns).Create(ctx, &v1alpha1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: rss,
			},
			Spec: v1alpha1.SubscriptionSpec{},
		}, metav1.CreateOptions{})

		profile.Spec.Subscription = v1.LocalObjectReference{
			Name: sub.Name,
		}
		clientset.Osf2fV1alpha1().Profiles(ns).Update(ctx, profile, metav1.UpdateOptions{})
	} else {
		sub, _ := clientset.Osf2fV1alpha1().Subscriptions(ns).Get(ctx, profile.Spec.Subscription.Name, metav1.GetOptions{})
		sub.Spec.RSSList = uniqueAppend(sub.Spec.RSSList, v1.LocalObjectReference{Name: rss})
		clientset.Osf2fV1alpha1().Subscriptions(ns).Update(ctx, sub, metav1.UpdateOptions{})
	}
	response.Write([]byte("ok"))
}

func (r Profile) unsubscribe(request *restful.Request, response *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	rss := request.QueryParameter(r.rssQuery.Data().Name)
	profileName := request.PathParameter(r.profilePath.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	if profile.Spec.Subscription.Name != "" {
		sub, _ := clientset.Osf2fV1alpha1().Subscriptions(ns).Get(ctx, profile.Spec.Subscription.Name, metav1.GetOptions{})
		var removed bool
		sub.Spec.RSSList, removed = removeLocalObjectReference(sub.Spec.RSSList, v1.LocalObjectReference{Name: rss})
		if removed {
			clientset.Osf2fV1alpha1().Subscriptions(ns).Update(ctx, sub, metav1.UpdateOptions{})
		}
	}
	response.Write([]byte("ok"))
}

func (r Profile) subscriptions(request *restful.Request, response *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	profileName := request.PathParameter(r.profilePath.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	var rssList []*v1alpha1.RSS
	if profile.Spec.Subscription.Name != "" {
		sub, _ := clientset.Osf2fV1alpha1().Subscriptions(ns).Get(ctx, profile.Spec.Subscription.Name, metav1.GetOptions{})
		for i := range sub.Spec.RSSList {
			rssNameRef := sub.Spec.RSSList[i]
			rss, _ := clientset.Osf2fV1alpha1().RSSes(ns).Get(ctx, rssNameRef.Name, metav1.GetOptions{})
			rssList = append(rssList, rss)
		}
	}
	response.WriteAsJson(rssList)
}

func (r Profile) playLater(req *restful.Request, resp *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	profileName := req.PathParameter(r.profilePath.Data().Name)
	episodeName := req.QueryParameter(r.episodeQuery.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	var added bool
	profile.Spec.LaterPlayList, added = uniquePlayToDoAppend(profile.Spec.LaterPlayList, v1alpha1.PlayTodo{
		LocalObjectReference: v1.LocalObjectReference{Name: episodeName},
	})
	if added {
		clientset.Osf2fV1alpha1().Profiles(ns).Update(ctx, profile, metav1.UpdateOptions{})
	}
}

func (r Profile) playOver(req *restful.Request, resp *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	profileName := req.PathParameter(r.profilePath.Data().Name)
	episodeName := req.QueryParameter(r.episodeQuery.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	var removed bool
	profile.Spec.LaterPlayList, removed = removePlayTodo(profile.Spec.LaterPlayList, v1alpha1.PlayTodo{
		LocalObjectReference: v1.LocalObjectReference{Name: episodeName},
	})
	if removed {
		clientset.Osf2fV1alpha1().Profiles(ns).Update(ctx, profile, metav1.UpdateOptions{})
	}
}

func (r Profile) notifier(req *restful.Request, resp *restful.Response) {
	config, err := clientcmd.BuildConfigFromFlags("", "/Users/rick/.kube/config")
	if err != nil {
		panic(err.Error())
	}
	profileName := req.PathParameter(r.profilePath.Data().Name)
	feishuWebhook := req.QueryParameter(r.feishuQuery.Data().Name)

	ctx := context.Background()
	clientset, err := client.NewForConfig(config)
	profile, err := clientset.Osf2fV1alpha1().Profiles(ns).Get(ctx, profileName, metav1.GetOptions{})

	if profile.Spec.Notifier.Name == "" {
		notifier := &v1alpha1.Notifier{}
		notifier.GenerateName = "auto"
		notifier.Spec.Feishu = &v1alpha1.FeishuNotifier{
			WebhookUrl: feishuWebhook,
		}

		notifier, _ = clientset.Osf2fV1alpha1().Notifiers(ns).Create(ctx, notifier, metav1.CreateOptions{})

		profile.Spec.Notifier = v1.LocalObjectReference{
			Name: notifier.Name,
		}
		clientset.Osf2fV1alpha1().Profiles(ns).Update(ctx, profile, metav1.UpdateOptions{})
	}
	resp.Write([]byte("ok"))
}

func removePlayTodo(todoList []v1alpha1.PlayTodo, todo v1alpha1.PlayTodo) (result []v1alpha1.PlayTodo, removed bool) {
	for i := range todoList {
		if todoList[i].Name == todo.Name {
			result = append(todoList[:i], todoList[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		result = todoList
	}
	return
}

func uniquePlayToDoAppend(todoList []v1alpha1.PlayTodo, todo v1alpha1.PlayTodo) (result []v1alpha1.PlayTodo, added bool) {
	found := false
	for i := range todoList {
		if todoList[i].Name == todo.Name {
			found = true
			break
		}
	}
	result = todoList
	if !found {
		result = append(result, todo)
		added = true
	}
	return
}

func removeLocalObjectReference(list []v1.LocalObjectReference, reference v1.LocalObjectReference) (
	result []v1.LocalObjectReference, removed bool) {
	for i := range list {
		if list[i].Name == reference.Name {
			result = append(list[:i], list[i+1:]...)
			removed = true
			break
		}
	}
	if !removed {
		result = list
	}
	return
}
